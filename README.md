# Blog Server API

A Go-based REST API for managing blog posts, user authentication, and image uploads. The project uses Gin for routing, MySQL for persistence, JWT-based authentication, and S3-compatible storage for article cover images.

## Features

- Create, list, retrieve, update, and delete blog articles
- User registration, login, token refresh, and password changes
- Image upload support for article cover images
- CORS support for local frontend development
- Structured project layout with separate layers for handlers, services, repositories, and DTOs

## Tech Stack

- Go
- Gin Web Framework
- MySQL
- JWT
- AWS S3 SDK (used with Backblaze B2-compatible storage)
- Godotenv

## Project Structure

- auth/: authentication handlers, services, DTOs, and repository logic
- blog/: blog article handlers, services, models, DTOs, and repository logic
- database/: database connection setup
- router/: route registration and middleware setup
- share/: shared utilities, JWT helpers, validation, and storage integration

## Prerequisites

- Go 1.25 or newer
- MySQL server
- A storage provider compatible with S3 APIs (such as Backblaze B2)

## Environment Variables

Create a `.env` file in the project root with values similar to:

```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=blog_db

B2_REGION=your-region
B2_ACCESS_KEY_ID=your-access-key
B2_SECRET_ACCESS_KEY=your-secret-key
B2_ENDPOINT=https://your-endpoint
B2_BUCKET=your-bucket-name
```

## Getting Started

1. Clone the repository:

```bash
git clone <repository-url>
cd blog_server
```

2. Install dependencies:

```bash
go mod tidy
```

3. Start the server:

```bash
go run .
```

The server will run on port `8080` by default.

## API Endpoints

### Articles

- `GET /api/articles` - List all articles
- `POST /api/articles` - Create a new article
- `GET /api/articles/:id` - Get one article by ID
- `PUT /api/articles/:id` - Update an article
- `DELETE /api/articles/:id` - Delete an article

### Authentication

- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `POST /api/auth/change-password`

## Testing

Run tests with:

```bash
go test ./...
```

## Notes

- Article creation and updates support multipart uploads for a `cover_image` field.
- CORS is configured for common local origins such as `http://localhost:5500` and `http://localhost:8000`.
