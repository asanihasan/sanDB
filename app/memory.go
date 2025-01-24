package app

import (
	"sort"
)

// Check and maintain max data length in inMemoryData
func MaintainMaxDataLength() {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	if len(lastAccessTimestamps) > AppConfig.Memory.MaxData {
		// Sort file paths by last access timestamps (oldest first)
		sortedPaths := make([]string, 0, len(lastAccessTimestamps))
		for path := range lastAccessTimestamps {
			sortedPaths = append(sortedPaths, path)
		}
		sort.Slice(sortedPaths, func(i, j int) bool {
			return lastAccessTimestamps[sortedPaths[i]] < lastAccessTimestamps[sortedPaths[j]]
		})

		// Remove oldest entries until length is within the limit
		for len(lastAccessTimestamps) > AppConfig.Memory.MaxData {
			oldestPath := sortedPaths[0]
			delete(inMemoryData, oldestPath)
			delete(lastAccessTimestamps, oldestPath)
			sortedPaths = sortedPaths[1:]
		}
	}
}

// Check and maintain max memory size of inMemoryData
func MaintainMaxMemorySize() {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	calculateMemorySize := func() int {
		totalSize := 0
		for _, fileData := range inMemoryData {
			for _, data := range fileData {
				totalSize += len(data) // Approximate size by string length
			}
		}
		return totalSize / (1024 * 1024) // Convert bytes to MB
	}

	currentSize := calculateMemorySize()

	if currentSize > AppConfig.Memory.MaxSize {
		// Sort file paths by last access timestamps (oldest first)
		sortedPaths := make([]string, 0, len(lastAccessTimestamps))
		for path := range lastAccessTimestamps {
			sortedPaths = append(sortedPaths, path)
		}
		sort.Slice(sortedPaths, func(i, j int) bool {
			return lastAccessTimestamps[sortedPaths[i]] < lastAccessTimestamps[sortedPaths[j]]
		})

		// Remove oldest entries until size is within the limit
		for currentSize > AppConfig.Memory.MaxSize {
			oldestPath := sortedPaths[0]

			// Calculate size to subtract
			if fileData, exists := inMemoryData[oldestPath]; exists {
				for _, data := range fileData {
					currentSize -= len(data) / (1024 * 1024) // Approximate size by string length
				}
			}

			delete(inMemoryData, oldestPath)
			delete(lastAccessTimestamps, oldestPath)
			sortedPaths = sortedPaths[1:]
		}
	}
}
