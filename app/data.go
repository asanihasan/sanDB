package app

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	inMemoryData        = make(map[string]map[int64][]byte) // File path -> Data map
	lastAccessTimestamps = make(map[string]int64)          // File path -> Last access timestamp
	dataMutex    sync.RWMutex                       // Mutex to handle concurrent access
)

func add_data(c *gin.Context) {
	dataPath := "./data" // Base directory for data
	collectionName := c.Param("collection_name") // Get collection name from path

	// Check if the collection directory exists
	collectionDir := fmt.Sprintf("%s/%s", dataPath, collectionName)
	if _, err := os.Stat(collectionDir); os.IsNotExist(err) {
		c.JSON(404, gin.H{"error": fmt.Sprintf("Collection '%s' does not exist", collectionName)})
		return
	}

	// Parse JSON body
	var requestData []struct {
		Time int64       `json:"time"` // Millisecond timestamp
		Data interface{} `json:"data"` // Data can be any JSON type
	}
	
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	for _, item := range requestData {
		if item.Time <= 0 {
			c.JSON(400, gin.H{"error": "Each item must have a valid 'time'"})
			return
		}

		// Convert `item.Data` to JSON byte slice
		dataJSON, err := json.Marshal(item.Data)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to process data: %v", err)})
			return
		}

		// Convert timestamp to year, day, and 6-hour segment
		t := time.UnixMilli(item.Time)
		year, day := t.Year(), t.YearDay()
		hour := t.Hour()
		segment := (hour / 6) + 1 // Calculate 6-hour segment (1-4)

		// Define file path
		segmentDir := fmt.Sprintf("%s/%d/%d", collectionDir, year, day)
		sanFilePath := fmt.Sprintf("%s/%d.san", segmentDir, segment)

		// Ensure directory exists
		if err := os.MkdirAll(segmentDir, os.ModePerm); err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create directory: %v", err)})
			return
		}

		// Safely load file into memory if not already loaded
		dataMutex.Lock()
		if _, exists := inMemoryData[sanFilePath]; !exists {
			inMemoryData[sanFilePath] = make(map[int64][]byte)

			// Load existing data if the file exists
			if _, err := os.Stat(sanFilePath); !os.IsNotExist(err) {
				file, err := os.Open(sanFilePath)
				if err != nil {
					dataMutex.Unlock()
					c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to open .san file: %v", err)})
					return
				}
				defer file.Close()

				// Temporary variable to decode data
				var tempData map[int64][]byte
				decoder := gob.NewDecoder(file)
				if err := decoder.Decode(&tempData); err != nil {
					dataMutex.Unlock()
					c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to decode .san file: %v", err)})
					return
				}
				inMemoryData[sanFilePath] = tempData
			}
		}

		// Insert or overwrite data in memory
		inMemoryData[sanFilePath][item.Time] = dataJSON

		// Update the last access timestamp
		lastAccessTimestamps[sanFilePath] = time.Now().Unix()

		dataMutex.Unlock()
	}

	// Save all modified files to disk
	for filePath, data := range inMemoryData {
		dataMutex.RLock() // Read lock for safe concurrent access
		file, err := os.Create(filePath)
		if err != nil {
			dataMutex.RUnlock()
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to save .san file: %v", err)})
			return
		}
		encoder := gob.NewEncoder(file)
		if err := encoder.Encode(data); err != nil {
			dataMutex.RUnlock()
			file.Close()
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to encode data to .san file: %v", err)})
			return
		}
		file.Close()
		dataMutex.RUnlock()
	}

	c.JSON(201, gin.H{"message": "Data added successfully"})
}

func get_data(c *gin.Context) {
	dataPath := "./data" // Base directory for data
	collectionName := c.Param("collection_name")
	collectionDir := fmt.Sprintf("%s/%s", dataPath, collectionName)

	// Check if the collection directory exists
	if _, err := os.Stat(collectionDir); os.IsNotExist(err) {
		c.JSON(404, gin.H{"error": fmt.Sprintf("Collection '%s' does not exist", collectionName)})
		return
	}

	// Parse query parameters
	start, err := strconv.ParseInt(c.Query("start"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid start parameter"})
		return
	}

	end, err := strconv.ParseInt(c.Query("end"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid end parameter"})
		return
	}

	limitParam := c.Query("limit")
	offsetParam := c.Query("offset")

	limit := -1
	offset := 0

	if limitParam != "" {
		limit, err = strconv.Atoi(limitParam)
		if err != nil || limit <= 0 {
			c.JSON(400, gin.H{"error": "Invalid limit parameter"})
			return
		}
	}

	if offsetParam != "" {
		offset, err = strconv.Atoi(offsetParam)
		if err != nil || offset < 0 {
			c.JSON(400, gin.H{"error": "Invalid offset parameter"})
			return
		}
	}

	// Convert start and end timestamps
	startTime := time.UnixMilli(start)
	endTime := time.UnixMilli(end)

	startYear, startDay := startTime.Year(), startTime.YearDay()
	startSegment := (startTime.Hour() / 6) + 1

	endYear, endDay := endTime.Year(), endTime.YearDay()
	endSegment := (endTime.Hour() / 6) + 1

	result := []map[string]interface{}{}

	// Loop through years
	for year := startYear; year <= endYear; year++ {
		yearDir := fmt.Sprintf("%s/%d", collectionDir, year)

		if _, err := os.Stat(yearDir); os.IsNotExist(err) {
			continue
		}

		// Determine day range
		dayStart := 1
		dayEnd := 366

		if year == startYear {
			dayStart = startDay
		}
		if year == endYear {
			dayEnd = endDay
		}

		for day := dayStart; day <= dayEnd; day++ {
			dayDir := fmt.Sprintf("%s/%d", yearDir, day)

			if _, err := os.Stat(dayDir); os.IsNotExist(err) {
				continue
			}

			files, err := os.ReadDir(dayDir)
			if err != nil {
				c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to read directory: %v", err)})
				return
			}

			for _, file := range files {
				if file.IsDir() {
					continue
				}

				segment, err := strconv.Atoi(file.Name()[:len(file.Name())-4])
				if err != nil {
					continue
				}

				if (year == startYear && day == startDay && segment < startSegment) ||
					(year == endYear && day == endDay && segment > endSegment) {
					continue
				}

				filePath := fmt.Sprintf("%s/%s", dayDir, file.Name())

				// Check if the file is already in memory
				dataMutex.RLock()
				fileData, exists := inMemoryData[filePath]
				dataMutex.RUnlock()

				if !exists {
					// Load the file into memory
					dataMutex.Lock()
					file, err := os.Open(filePath)
					if err != nil {
						dataMutex.Unlock()
						c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to open file: %v", err)})
						return
					}

					defer file.Close()

					fileData = make(map[int64][]byte)
					decoder := gob.NewDecoder(file)
					if err := decoder.Decode(&fileData); err != nil {
						dataMutex.Unlock()
						c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to decode file: %v", err)})
						return
					}

					inMemoryData[filePath] = fileData
					dataMutex.Unlock()
				}

				// Filter data by timestamp range
				for ts, data := range fileData {
					if ts >= start && ts <= end {
						var deserializedData interface{}

						// Attempt to unmarshal the data into a generic interface{}
						err := json.Unmarshal(data, &deserializedData)
						if err != nil {
							// If unmarshaling fails, keep the original data as is
							deserializedData = string(data)
						}

						result = append(result, map[string]interface{}{
							"time": ts,
							"data": deserializedData,
						})
					}
				}
			}
		}
	}

	// Sort results by timestamp
	sort.Slice(result, func(i, j int) bool {
		return result[i]["time"].(int64) < result[j]["time"].(int64)
	})

	// Apply offset and limit
	if offset < len(result) {
		result = result[offset:]
	}
	if limit > 0 && limit < len(result) {
		result = result[:limit]
	}

	c.JSON(200, gin.H{"data": result})
}

func delete_data(c *gin.Context) {
	dataPath := "./data" // Base directory for data
	collectionName := c.Param("collection_name")
	collectionDir := fmt.Sprintf("%s/%s", dataPath, collectionName)

	// Check if the collection directory exists
	if _, err := os.Stat(collectionDir); os.IsNotExist(err) {
		c.JSON(404, gin.H{"error": fmt.Sprintf("Collection '%s' does not exist", collectionName)})
		return
	}

	// Parse query parameters
	start, err := strconv.ParseInt(c.Query("start"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid start parameter"})
		return
	}

	end, err := strconv.ParseInt(c.Query("end"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid end parameter"})
		return
	}

	// Convert start and end timestamps
	startTime := time.UnixMilli(start)
	endTime := time.UnixMilli(end)

	startYear, startDay := startTime.Year(), startTime.YearDay()
	startSegment := (startTime.Hour() / 6) + 1

	endYear, endDay := endTime.Year(), endTime.YearDay()
	endSegment := (endTime.Hour() / 6) + 1

	// Loop through years
	for year := startYear; year <= endYear; year++ {
		yearDir := fmt.Sprintf("%s/%d", collectionDir, year)

		if _, err := os.Stat(yearDir); os.IsNotExist(err) {
			continue
		}

		// Determine day range
		dayStart := 1
		dayEnd := 366

		if year == startYear {
			dayStart = startDay
		}
		if year == endYear {
			dayEnd = endDay
		}

		for day := dayStart; day <= dayEnd; day++ {
			dayDir := fmt.Sprintf("%s/%d", yearDir, day)

			if _, err := os.Stat(dayDir); os.IsNotExist(err) {
				continue
			}

			files, err := os.ReadDir(dayDir)
			if err != nil {
				c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to read directory: %v", err)})
				return
			}

			for _, file := range files {
				if file.IsDir() {
					continue
				}

				segment, err := strconv.Atoi(file.Name()[:len(file.Name())-4])
				if err != nil {
					continue
				}

				filePath := fmt.Sprintf("%s/%s", dayDir, file.Name())

				if (year == startYear && day == startDay && segment < startSegment) ||
					(year == endYear && day == endDay && segment > endSegment) {
					continue
				}

				// Check if the file is already in memory
				dataMutex.Lock()
				fileData, exists := inMemoryData[filePath]
				if !exists {
					// Load the file into memory
					file, err := os.Open(filePath)
					if err != nil {
						dataMutex.Unlock()
						c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to open file: %v", err)})
						return
					}

					defer file.Close()

					fileData = make(map[int64][]byte)
					decoder := gob.NewDecoder(file)
					if err := decoder.Decode(&fileData); err != nil {
						dataMutex.Unlock()
						c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to decode file: %v", err)})
						return
					}
					inMemoryData[filePath] = fileData
				}

				// Remove data inside the range from inMemoryData
				for ts := range fileData {
					if ts >= start && ts <= end {
						delete(fileData, ts)
					}
				}

				// Rewrite the file if data remains, otherwise delete the file
				if len(fileData) > 0 {
					file, err := os.Create(filePath)
					if err != nil {
						dataMutex.Unlock()
						c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to rewrite file: %v", err)})
						return
					}
					defer file.Close()

					encoder := gob.NewEncoder(file)
					if err := encoder.Encode(fileData); err != nil {
						dataMutex.Unlock()
						c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to encode file: %v", err)})
						return
					}
				} else {
					if err := os.Remove(filePath); err != nil {
						dataMutex.Unlock()
						c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to delete file: %v", err)})
						return
					}
					delete(inMemoryData, filePath) // Remove from inMemoryData
					delete(lastAccessTimestamps, filePath) // Remove from lastAccessTimestamps
				}
				dataMutex.Unlock()
			}
		}
	}

	c.JSON(200, gin.H{"message": "Data deleted successfully"})
}
