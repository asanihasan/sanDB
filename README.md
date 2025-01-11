# SanDB

SanDB is a lightweight binary data storage backend written in Go, using the Gin framework for API management. It provides an organized directory structure and API endpoints to manage collections of binary data.

---

## Features

- **Dynamic Collection Management**:

  - Create, read, update, and delete collections.
  - Organize collections under a `data` directory.

- **Authorization**:

  - Uses token-based authorization to secure API endpoints.

- **API-Driven**:
  - RESTful API endpoints for seamless integration with other applications.

---

## Project Structure

```
SanDB/
├── app/
│   ├── server.go         # Web server setup
│   ├── data.go           # Data-related routes
│   ├── config.go         # Configuration loader
│   ├── collections.go    # Collection-related routes
├── config/
│   ├── config.yml        # Configuration file
├── data/                 # Directory for collections
├── main.go               # Entry point
├── go.mod                # Go module file
```

---

## Configuration

Edit the `config/config.yml` file to specify the server port and authorization token:

```yaml
server:
  port: 6969
  token: "your-secret-token"
```

---

## API Endpoints

**Base URL**: `http://localhost:<PORT>/`

### **Collections**

1. **List Collections**

   - **Endpoint**: `GET /collections`
   - **Description**: Retrieves a list of all collections.
   - **Response**:
     - `200 OK` : Return list of collections
     ```json
     {
       "collections": ["collection1", "collection2"]
     }
     ```

2. **Get Collection Details**

   - **Endpoint**: `GET /collections/:collection_name`
   - **Description**: Checks if a collection exists.
   - **Response**:
     - `200 OK` : Collection 'collection_name' exists
     - `404 Not Found` : Collection 'collection_name' does not exist

3. **Create a Collection**

   - **Endpoint**: `PUT /collections/:collection_name`
   - **Description**: Creates a new collection.
   - **Response**:
     - `201 Created` : Collection 'collection_name' created
     - `409 Conflict` : Collection 'collection_name' already exists

4. **Delete a Collection**

   - **Endpoint**: `DELETE /collections/:collection_name`
   - **Description**: Deletes a collection.
   - **Response**:
     - `200 OK` : Collection 'collection_name' deleted successfully
     - `404 Not Found` : Collection 'collection_name' does not exist

5. **Rename a Collection**

   - **Endpoint**: `PATCH /collections/:collection_name?new_name=new`
   - **Description**: Renames an existing collection.
   - **Response**:
     - `200 OK` : Collection 'old' renamed to 'new'
     - `404 Not Found` : Collection 'old' does not exist
     - `409 Conflict` : Collection 'new' already exists

### **Data**

1. **Add Data**
    - **Endpoint**: `PUT /data/:collection_name`

    - **Description**: Adds data to the specified collection.
    - **Parameters**:
      - `:collection_name` (path): Name of the collection to add data to.
    - **Request Body** (JSON Array):
      ```json
      [
        {
          "time": 1672531200000, // Millisecond timestamp
          "data": "example data" // Data (can be any JSON type)
        },
        {
          "time": 1672534800000,
          "data": { "key": "value" }
        }
      ]
      ```
    - **Response**:
      - `201 Created`: Data added successfully.
      - `400 Bad Request`: Invalid input or missing fields.
      - `404 Not Found`: Collection does not exist.
      - `500 Internal Server Error`: Server-side error.

2. **Retrieve Data**
    - **Endpoint**: `GET /data/:collection_name`

    - **Description**: Retrieves data from the specified collection within a time range.
    - **Parameters**:
      - `:collection_name` (path): Name of the collection to retrieve data from.
      - `start` (query): Start time in milliseconds (required).
      - `end` (query): End time in milliseconds (required).
      - `limit` (query): Maximum number of records to return (optional).
      - `offset` (query): Number of records to skip (optional).
    - **Response**:
      - `200 OK`: Returns a JSON array of data points.
        ```json
        {
          "data": [
            {
              "time": 1672531200000,
              "data": "example data"
            },
            {
              "time": 1672534800000,
              "data": { "key": "value" }
            }
          ]
        }
        ```
      - `400 Bad Request`: Missing or invalid query parameters.
      - `404 Not Found`: Collection does not exist.
      - `500 Internal Server Error`: Server-side error.

3. **Delete Data**
    - **Endpoint**: `DELETE /data/:collection_name`

    - **Description**: Deletes data in the specified collection within a time range.
    - **Parameters**:
      - `:collection_name` (path): Name of the collection to delete data from.
      - `start` (query): Start time in milliseconds (required).
      - `end` (query): End time in milliseconds (required).
    - **Response**:
      - `200 OK`: Data deleted successfully.
      - `400 Bad Request`: Missing or invalid query parameters.
      - `404 Not Found`: Collection does not exist.
      - `500 Internal Server Error`: Server-side error.

---

## **Example Usage**

### Add Data Example
```bash
curl -X PUT http://localhost:6969/data/my_collection \
-H "Content-Type: application/json" \
-d '[
  {"time": 1672531200000, "data": "example data"},
  {"time": 1672534800000, "data": {"key": "value"}}
]'
```

### Retrieve Data Example
```bash
curl -X GET "http://localhost:6969/data/my_collection?start=1672531200000&end=1672538400000&limit=10&offset=0"
```

### Delete Data Example
```bash
curl -X DELETE "http://localhost:6969/data/my_collection?start=1672531200000&end=1672538400000"
```

---

## Notes
- **Time Range**: Timestamps must be in milliseconds (Unix epoch format).
- **Collections**: Collections must exist before adding, retrieving, or deleting data.
- **Error Handling**: Ensure proper handling of API responses to manage errors effectively.

---

## Authorization

All API requests require an `Authorization` header containing the token specified in `config.yml`.

Example:

```bash
curl -H "Authorization: your-secret-token" http://localhost:6969/collection
```

---

## Running the Project

1. **Clone Repository**:

   ```bash
   git clone https://github.com/asanihasan/sanDB.git
   ```

2. **init go mod**:

   ```bash
   go mod init sanDB
   ```

3. **Install Dependencies**:

   ```bash
   go mod tidy
   ```

4. **Run the Server**:

   ```bash
   go run main.go
   ```

5. **Access the API**:
   Visit `http://localhost:6969/`.

---

## Contributions

Feel free to fork and contribute by submitting pull requests. Suggestions and feedback are welcome.

---
