# Flyhalf

### A Very, Very Opinionated Agile Scrum Management System

Full-stack ticketing, epic, and sprint management application with Go API, vanilla JavaScript SPA, PostgreSQL database, and JWT authentication with role-based authorization.

## Tech Stack

- **Backend**: Go 1.24 with chi router, pgx (PostgreSQL driver), golang-jwt
- **Frontend**: Vanilla JavaScript SPA with ES modules (no build step required)
- **Database**: PostgreSQL 16
- **Authentication**: JWT access tokens (15min) + refresh tokens (7 days, HttpOnly cookie)
- **Development**: Docker Compose with hot reload (Air for Go)

## Features

- JWT-based authentication with token refresh
- Role-based access control (admin/user)
- **Ticket Management**:
  - CRUD operations with 6 status options (new, open, in-progress, blocked, needs-review, closed)
  - New tickets automatically default to "new" status
  - New tickets highlighted with blue background and sorted to top
  - Priority system with "Promote to Top" button to bump tickets to top of list
  - Ticket assignment to users
  - Assign tickets to epics for organization
  - Assign tickets to sprints for sprint planning
  - 6-character unique ID for each ticket
- **Epic Management**:
  - CRUD operations for epics (name and description)
  - Organize tickets by assigning them to epics
  - Epic detail view shows all tickets assigned to that epic
  - Full list and detail views
- **Sprint Management**:
  - CRUD operations for sprints (name and start date)
  - End date automatically calculated as 2 weeks after start date
  - Assign tickets to sprints for sprint planning
  - Sprint detail view shows all tickets assigned to that sprint
  - Full list and detail views
- All users can view and edit all tickets and epics (collaborative workspace)
- Users can delete tickets/epics they created; admins can delete any ticket/epic
- Forced password change for newly created users
- Admin user management
- User settings page with account information
- Password change functionality
- Responsive UI with modern CSS
- Client-side routing with hash-based navigation that preserves state on refresh
- Secure HttpOnly cookies for refresh tokens

## Permission Model

### Regular Users (role: 'user')
- ✅ View all tickets, epics, and sprints
- ✅ Create new tickets, epics, and sprints
- ✅ Edit any ticket, epic, or sprint
- ✅ Delete tickets, epics, and sprints they created
- ✅ Assign tickets to epics and sprints
- ✅ Promote tickets to top of list
- ✅ Change own password
- ✅ View own account settings
- ❌ Delete tickets/epics/sprints created by others
- ❌ Manage users

### Administrators (role: 'admin')
- ✅ All user permissions
- ✅ Delete any ticket, epic, or sprint (including those created by others)
- ✅ Create new users (with forced password change)
- ✅ Edit user accounts
- ✅ Delete users
- ✅ Deactivate/activate users

This collaborative permission model allows all team members to view and update tickets, epics, and sprints while protecting data integrity. Users can manage their own items completely, but cannot delete items created by others.

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

## User Interface

The application provides the following pages:

### For All Users
- **Login Page** - Email/password authentication
- **Force Password Change** - Required for newly created users on first login
- **Tickets List** - View all tickets with 6-character ID, title, status badges, epic, and sprint
  - New tickets highlighted with blue background
  - New tickets sorted to top of list
  - "Promote to Top" button to bump tickets to top priority
- **Ticket Detail** - View full ticket information including epic and sprint assignment with delete button (enabled only for own tickets)
- **Create/Edit Ticket** - Form to create or modify tickets
  - Create: Title and description only (status defaults to "new")
  - Edit: Additional fields for status (5 options), epic assignment, and sprint assignment
- **Epics List** - View all epics with name column
- **Epic Detail** - View epic name and description with table of all tickets assigned to the epic
- **Create/Edit Epic** - Form to create or modify epics (name and description)
- **Sprints List** - View all sprints with name, start date, and end date columns
- **Sprint Detail** - View sprint dates with table of all tickets assigned to the sprint
- **Create/Edit Sprint** - Form to create or modify sprints (name and start date, end date auto-calculated)
- **Settings** - View account information and change password

### Admin Only
- **User Management** - List all users
- **Create/Edit User** - Manage user accounts (new users must change password on first login)
- **User Detail** - View user information
- **Delete Users** - Remove user accounts
- **Delete Any Ticket/Epic** - Delete button enabled for all tickets and epics

### Navigation
- Click the **Flyhalf** logo to return to the tickets list
- Click your **username** in the navbar to access settings
- **Tickets** link shows all tickets
- **Epics** link shows all epics
- **Sprints** link shows all sprints
- **Users** link (admins only) for user management
- **Logout** button to end session
- Page state preserved on browser refresh

### New User Workflow
1. Admin creates user with temporary password
2. User receives credentials and logs in
3. **Immediately redirected** to forced password change page
4. Must change password before accessing application
5. After password change, redirected to tickets page

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
| PUT | `/auth/password` | Change password | Yes |

### Ticket Endpoints

| Method | Endpoint | Description | Auth Required | Role |
|--------|----------|-------------|---------------|------|
| GET | `/tickets` | List all tickets | Yes | Any |
| POST | `/tickets` | Create ticket | Yes | Any |
| GET | `/tickets/{id}` | Get ticket by ID | Yes | Any |
| PUT | `/tickets/{id}` | Update ticket | Yes | Any |
| DELETE | `/tickets/{id}` | Delete ticket | Yes | Creator or Admin |
| POST | `/tickets/{id}/promote` | Promote ticket to top | Yes | Any |

**Note**: All authenticated users can view and edit all tickets. Users can delete tickets they created; admins can delete any ticket.

**Ticket Fields**:
- `title` (string, required)
- `description` (string, optional)
- `status` (string: new, open, in-progress, blocked, needs-review, closed)
- `epic_id` (UUID, optional) - Assign ticket to an epic
- `sprint_id` (UUID, optional) - Assign ticket to a sprint
- `priority` (integer) - Automatically managed by promote feature

### Epic Endpoints

| Method | Endpoint | Description | Auth Required | Role |
|--------|----------|-------------|---------------|------|
| GET | `/epics` | List all epics | Yes | Any |
| POST | `/epics` | Create epic | Yes | Any |
| GET | `/epics/{id}` | Get epic by ID | Yes | Any |
| PUT | `/epics/{id}` | Update epic | Yes | Any |
| DELETE | `/epics/{id}` | Delete epic | Yes | Creator or Admin |

**Note**: All authenticated users can view and edit all epics. Users can delete epics they created; admins can delete any epic.

**Epic Fields**:
- `name` (string, required)
- `description` (string, optional)

### Sprint Endpoints

| Method | Endpoint | Description | Auth Required | Role |
|--------|----------|-------------|---------------|------|
| GET | `/sprints` | List all sprints | Yes | Any |
| POST | `/sprints` | Create sprint | Yes | Any |
| GET | `/sprints/{id}` | Get sprint by ID | Yes | Any |
| PUT | `/sprints/{id}` | Update sprint | Yes | Any |
| DELETE | `/sprints/{id}` | Delete sprint | Yes | Creator or Admin |

**Note**: All authenticated users can view and edit all sprints. Users can delete sprints they created; admins can delete any sprint.

**Sprint Fields**:
- `name` (string, required)
- `start_date` (date, required, format: YYYY-MM-DD)
- `end_date` (date, auto-calculated as start_date + 14 days)

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
- `must_change_password` (boolean, default: false)
- `created_at`, `updated_at` (timestamps)

### Refresh Tokens Table
- `id` (UUID, primary key)
- `user_id` (FK to users)
- `token_hash` (not null)
- `expires_at` (timestamp)
- `revoked_at` (timestamp, nullable)
- `created_at` (timestamp)

### Epics Table
- `id` (UUID, primary key)
- `user_id` (FK to users - epic creator)
- `name` (varchar(255), not null)
- `description` (text, nullable)
- `created_at`, `updated_at` (timestamps)

### Sprints Table
- `id` (UUID, primary key)
- `user_id` (FK to users - sprint creator)
- `name` (varchar(255), not null)
- `start_date` (date, not null)
- `end_date` (date, not null)
- `created_at`, `updated_at` (timestamps)

### Tickets Table
- `id` (UUID, primary key)
- `user_id` (FK to users - ticket creator)
- `title` (varchar(255), not null)
- `description` (text, nullable)
- `status` (varchar(50): new, open, in-progress, blocked, needs-review, closed, default: 'new')
- `assigned_to` (UUID, FK to users, nullable)
- `epic_id` (UUID, FK to epics, nullable)
- `sprint_id` (UUID, FK to sprints, nullable)
- `priority` (integer, default: 0)
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
