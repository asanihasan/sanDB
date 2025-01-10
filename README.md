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
│   ├── config.go         # Configuration loader
│   ├── api.go            # Collection-related routes
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

### **Base URL**: `http://localhost:<PORT>/`

1. **List Collections**

   - **Endpoint**: `GET /collection`
   - **Description**: Retrieves a list of all collections.
   - **Response**:
     ```json
     {
       "collections": ["collection1", "collection2"]
     }
     ```

2. **Get Collection Details**

   - **Endpoint**: `GET /collections/:collection_name`
   - **Description**: Checks if a collection exists.
   - **Response**:
     ```json
     {
       "message": "Collection 'collection_name' exists"
     }
     ```

3. **Create a Collection**

   - **Endpoint**: `PUT /collections/:collection_name`
   - **Description**: Creates a new collection.
   - **Response**:
     ```json
     {
       "message": "Collection 'collection_name' created"
     }
     ```

4. **Delete a Collection**

   - **Endpoint**: `DELETE /collections/:collection_name`
   - **Description**: Deletes a collection.
   - **Response**:
     ```json
     {
       "message": "Collection 'collection_name' deleted successfully"
     }
     ```

5. **Rename a Collection**
   - **Endpoint**: `PATCH /collections?old_name=old&new_name=new`
   - **Description**: Renames an existing collection.
   - **Response**:
     ```json
     {
       "message": "Collection 'old' renamed to 'new'"
     }
     ```

---

## Authorization

All API requests require an `Authorization` header containing the token specified in `config.yml`.

Example:

```bash
curl -H "Authorization: your-secret-token" http://localhost:6969/collection
```

---

## Running the Project

1. **Install Dependencies**:

   ```bash
   go mod tidy
   ```

2. **Run the Server**:

   ```bash
   go run main.go
   ```

3. **Access the API**:
   Visit `http://localhost:6969/`.

---

## Contributions

Feel free to fork and contribute by submitting pull requests. Suggestions and feedback are welcome.

---
