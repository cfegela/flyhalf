# Flyhalf

Full-stack ticketing and task management application with Go API, vanilla JavaScript SPA, PostgreSQL database, and JWT authentication with role-based authorization.

## Tech Stack

- **Backend**: Go 1.23 with chi router, pgx (PostgreSQL driver), golang-jwt
- **Frontend**: Vanilla JavaScript SPA with ES modules (no build step required)
- **Database**: PostgreSQL 16
- **Authentication**: JWT access tokens (15min) + refresh tokens (7 days, HttpOnly cookie)
- **Development**: Docker Compose with hot reload (Air for Go)

## Features

- JWT-based authentication with token refresh
- Role-based access control (admin/user)
- CRUD operations for tickets with status and priority tracking
- Admin user management
- Ticket assignment and priority management
- Responsive UI with modern CSS
- Toast notifications
- Client-side routing
- Secure HttpOnly cookies for refresh tokens

## Project Structure

```
flyhalf/
├── api/                    # Go backend
│   ├── cmd/server/         # Application entry point
│   ├── internal/           # Private application code
│   │   ├── auth/          # Authentication & JWT
│   │   ├── config/        # Configuration
│   │   ├── database/      # Database connection & migrations
│   │   ├── handler/       # HTTP handlers
│   │   ├── middleware/    # HTTP middleware
│   │   ├── model/         # Data models & repositories
│   │   └── router/        # Route definitions
│   └── Dockerfile
├── web/                    # JavaScript frontend
│   ├── css/               # Stylesheets
│   ├── js/                # JavaScript modules
│   │   ├── components/    # UI components
│   │   └── views/         # Page views
│   └── nginx.conf
├── scripts/               # Utility scripts
└── docker-compose.yml
```

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Git

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd flyhalf
```

2. The `.env` file has already been created with secure JWT secrets. Review and modify if needed:
```bash
cat .env
```

3. Start the application:
```bash
docker-compose up
```

This will start three services:
- **PostgreSQL** on port 5432
- **API** on port 8081
- **Web** on port 3000

4. Wait for the services to start. You should see:
```
flyhalf-api    | Server starting on port 8080
```

### Creating the Initial Admin User

The application automatically runs database migrations on startup. To create an initial admin user:

1. Connect to the PostgreSQL container:
```bash
docker exec -it flyhalf-db psql -U flyhalf -d flyhalf
```

2. Run the seed script:
```bash
\i /scripts/create-admin.sql
```

Or alternatively, from your host machine:
```bash
docker exec -i flyhalf-db psql -U flyhalf -d flyhalf < scripts/create-admin.sql
```

**Default Admin Credentials:**
- Email: `admin@flyhalf.local`
- Password: `admin123`

**IMPORTANT**: Change this password immediately after first login!

### Accessing the Application

Open your browser and navigate to:
```
http://localhost:3000
```

Log in with the admin credentials above.

## API Documentation

### Base URL
```
http://localhost:8081/api/v1
```

### Authentication Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/auth/login` | Login with email/password | No |
| POST | `/auth/refresh` | Refresh access token | No (requires refresh token cookie) |
| POST | `/auth/logout` | Logout and revoke tokens | Yes |
| GET | `/auth/me` | Get current user info | Yes |

### Ticket Endpoints

| Method | Endpoint | Description | Auth Required | Role |
|--------|----------|-------------|---------------|------|
| GET | `/tickets` | List tickets | Yes | Any |
| POST | `/tickets` | Create ticket | Yes | Any |
| GET | `/tickets/{id}` | Get ticket by ID | Yes | Any |
| PUT | `/tickets/{id}` | Update ticket | Yes | Owner or Admin |
| DELETE | `/tickets/{id}` | Delete ticket | Yes | Owner or Admin |

### Admin Endpoints

| Method | Endpoint | Description | Auth Required | Role |
|--------|----------|-------------|---------------|------|
| GET | `/admin/users` | List all users | Yes | Admin |
| POST | `/admin/users` | Create user | Yes | Admin |
| GET | `/admin/users/{id}` | Get user by ID | Yes | Admin |
| PUT | `/admin/users/{id}` | Update user | Yes | Admin |
| DELETE | `/admin/users/{id}` | Delete user | Yes | Admin |

### Authentication Header

For authenticated requests, include the JWT access token in the Authorization header:
```
Authorization: Bearer <access_token>
```

## Development

### Hot Reload

The development environment uses Air for Go hot reload. Any changes to Go files will automatically rebuild and restart the server.

For frontend changes, simply refresh your browser - no build step required!

### Running Tests

```bash
cd api
go test ./...
```

### Database Migrations

Migrations run automatically on application startup. The migration code is in:
```
api/internal/database/database.go
```

### Adding New Dependencies

Go:
```bash
cd api
go get <package>
go mod tidy
```

Frontend: No package manager needed - just add ES module imports!

## Database Schema

### Users Table
- `id` (UUID, primary key)
- `email` (unique, not null)
- `password_hash` (not null)
- `role` (enum: 'admin', 'user')
- `first_name` (not null)
- `last_name` (not null)
- `is_active` (boolean)
- `created_at`, `updated_at` (timestamps)

### Refresh Tokens Table
- `id` (UUID, primary key)
- `user_id` (FK to users)
- `token_hash` (not null)
- `expires_at` (timestamp)
- `revoked_at` (timestamp, nullable)
- `created_at` (timestamp)

### Tickets Table
- `id` (UUID, primary key)
- `user_id` (FK to users - ticket creator)
- `title` (not null)
- `description` (text)
- `status` (varchar: open, in_progress, resolved, closed)
- `priority` (varchar: low, medium, high, urgent)
- `assigned_to` (UUID, FK to users, nullable)
- `metadata` (JSONB)
- `created_at`, `updated_at` (timestamps)

## Security

- Passwords hashed with bcrypt (cost 12)
- Access tokens: Short-lived (15 minutes), stored in memory only
- Refresh tokens: HttpOnly + Secure + SameSite=Strict cookies
- CORS configured with explicit origin allowlist
- Parameterized queries to prevent SQL injection
- Security headers (X-Content-Type-Options, X-Frame-Options, etc.)

## Production Deployment

1. Update `.env` with production values:
   - Generate new JWT secrets: `openssl rand -base64 32`
   - Set `ENVIRONMENT=production`
   - Configure proper database credentials
   - Set `DB_SSLMODE=require`
   - Update `ALLOWED_ORIGIN` to your production domain

2. Use production Docker target:
```bash
docker-compose -f docker-compose.prod.yml up -d
```

3. Set up HTTPS with a reverse proxy (nginx/Caddy)

4. Regular backups of PostgreSQL database

## Troubleshooting

### Port Already in Use

If ports 3000, 5432, or 8081 are already in use, modify the port mappings in `docker-compose.yml`.

### Database Connection Issues

Check that PostgreSQL is healthy:
```bash
docker-compose ps
```

View logs:
```bash
docker-compose logs postgres
docker-compose logs api
```

### Frontend Not Loading

Check nginx logs:
```bash
docker-compose logs web
```

Ensure all JavaScript files are being served correctly.

## License

MIT

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.
