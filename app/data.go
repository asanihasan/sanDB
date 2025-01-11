package app

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var inMemoryData = make(map[string]map[int64]string) // File path -> Data map

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

		// Convert `item.Data` to JSON string
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

		// Load file into memory if not already loaded
		if _, exists := inMemoryData[sanFilePath]; !exists {
			inMemoryData[sanFilePath] = make(map[int64]string)

			// Load existing data if the file exists
			if _, err := os.Stat(sanFilePath); !os.IsNotExist(err) {
				file, err := os.Open(sanFilePath)
				if err != nil {
					c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to open .san file: %v", err)})
					return
				}
				// Explicitly close the file after reading
				func() {
					defer file.Close()

					var tempData map[int64]string
					decoder := gob.NewDecoder(file)
					if err := decoder.Decode(&tempData); err != nil {
						c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to decode .san file: %v", err)})
						return
					}
					inMemoryData[sanFilePath] = tempData
				}()
			}
		}

		// Insert or overwrite data in memory
		inMemoryData[sanFilePath][item.Time] = string(dataJSON)
	}

	// Save all modified files to disk
	for filePath, data := range inMemoryData {
		file, err := os.Create(filePath)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to save .san file: %v", err)})
			return
		}
		func() {
			defer file.Close()

			encoder := gob.NewEncoder(file)
			if err := encoder.Encode(data); err != nil {
				c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to encode data to .san file: %v", err)})
				return
			}
		}()
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
				fileData := make(map[int64]string)
			
				// Open the file
				f, err := os.Open(filePath)
				if err != nil {
					c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to open file: %v", err)})
					return
				}
			
				// Decode the file and immediately close it after processing
				func() {
					defer f.Close()
			
					decoder := gob.NewDecoder(f)
					if err := decoder.Decode(&fileData); err != nil {
						c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to decode file: %v", err)})
						return
					}
				}()
			
				// Process the file data
				for ts, data := range fileData {
					if ts >= start && ts <= end {
						var deserializedData interface{}
			
						// Attempt to unmarshal the data into a generic interface{}
						err := json.Unmarshal([]byte(data), &deserializedData)
						if err != nil {
							// If unmarshaling fails, keep the original data as is
							deserializedData = data
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
	endYear, endDay := endTime.Year(), endTime.YearDay()

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

				filePath := fmt.Sprintf("%s/%s", dayDir, file.Name())

				if (year == startYear && day == startDay) || (year == endYear && day == endDay) {
					// Edge file: Modify contents
					fileData := make(map[int64]string)
					f, err := os.Open(filePath)
					if err != nil {
						c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to open file: %v", err)})
						return
					}
					defer f.Close()

					decoder := gob.NewDecoder(f)
					if err := decoder.Decode(&fileData); err != nil {
						c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to decode file: %v", err)})
						return
					}

					// Remove data inside the range
					for ts := range fileData {
						if ts >= start && ts <= end {
							delete(fileData, ts)
						}
					}

					// Rewrite the file
					f, err = os.Create(filePath)
					if err != nil {
						c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to rewrite file: %v", err)})
						return
					}
					defer f.Close()

					encoder := gob.NewEncoder(f)
					if err := encoder.Encode(fileData); err != nil {
						c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to encode file: %v", err)})
						return
					}
				} else {
					// Delete the file
					if err := os.Remove(filePath); err != nil {
						c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to delete file: %v", err)})
						return
					}
				}
			}

			// Remove empty day directory if not start or end day
			if day != startDay && day != endDay {
				if err := os.Remove(dayDir); err == nil {
					continue
				}
			}
		}
	}

	c.JSON(200, gin.H{"message": "Data deleted successfully"})
}
