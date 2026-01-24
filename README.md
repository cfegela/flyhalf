# Flyhalf

### A Very, Very Opinionated Agile Scrum Management System

**flyhalf /ˈflaɪˌhɑːf/ noun**

The flyhalf is rugby’s primary playmaker and tactical leader who directs the team's attack. Flyhalves possess exceptional vision, game management, accuracy, and resilience.

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
  - CRUD operations with 5 status options (open, in-progress, blocked, needs-review, closed)
  - New tickets automatically default to "open" status
  - New tickets appear at bottom of list (priority 0, promoted tickets at top)
  - Priority system with "Promote to Top" button to bump tickets to top of list
  - Optional ticket sizing (Small=1, Medium=2, Large=3, X-Large=5, Danger=8)
  - Ticket assignment to users with assignee display in ticket list
  - Assign tickets to epics for organization (epic names shown as acronyms in list)
  - Assign tickets to sprints for sprint planning
  - Required title and description fields
  - 6-character unique ID for each ticket
  - Simplified list view with detail-level actions (edit/delete available in detail view only)
- **Epic Management**:
  - CRUD operations for epics with required name and description fields
  - Organize tickets by assigning them to epics
  - Epic detail view shows all tickets assigned to that epic
  - Epic names displayed as acronyms in ticket list (uppercase letters only)
  - Full list and detail views
  - Simplified list view with detail-level actions (edit/delete available in detail view only)
- **Sprint Management**:
  - CRUD operations for sprints (name and start date)
  - End date automatically calculated as 2 weeks after start date
  - Assign tickets to sprints for sprint planning
  - Sprint detail view shows all tickets assigned to that sprint
  - **Sprint Board**: Interactive kanban board with drag-and-drop functionality
    - Three columns: Committed (open), Underway (in-progress/blocked/needs-review), Completed (closed)
    - Drag tickets between columns to update their status
    - Clickable status badges in Underway column to change between in-progress/blocked/needs-review
    - Real-time status updates via API
    - Tickets sorted by priority within each column
    - Responsive design with mobile support
  - Full list and detail views
  - Simplified list view with detail-level actions (edit/delete available in detail view only)
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
- ✅ Use sprint board with drag-and-drop to update ticket status
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

## Recent UI/UX Improvements

The application has been enhanced with a comprehensive redesign focused on legibility and information hierarchy:

### Structured Card Layouts
All detail pages and forms now use organized card-based layouts with clear section headers:

**Detail Pages**:
- **Ticket Details**: Organized into Key Information, Description, Project Details, and Metadata cards
- **Epic Details**: Organized into Epic Details (with acronym display) and Tickets cards
- **Sprint Details**: Organized into Sprint Details (with status calculation), Timeline, and Tickets cards
  - Shows active/completed/upcoming status with color-coded badges
  - Displays duration and days remaining/until start
- **User Details**: Organized into User Information and Access & Permissions cards
- **Settings Page**: Organized into Account Information and Security cards

**Create/Edit Forms**:
- **Ticket Form**: Organized into Basic Information, Assignment & Sizing, Project Organization (edit only), and Form Actions cards
  - Responsive 2-column grids for related fields
  - Helpful placeholders and contextual hints
- **Epic Form**: Organized into Epic Information and Form Actions cards
  - Guidance about acronym generation from uppercase letters
- **Sprint Form**: Organized into Sprint Information and Form Actions cards
  - Clear explanation of automatic end date calculation
- **User Form**: Organized into Personal Information, Security (create only), Access & Permissions, and Form Actions cards
  - 2-column grid for name fields
  - Explanations for role permissions and account status

### Visual Enhancements
- **Improved Typography**: Larger, more prominent text for key information with consistent font sizes and weights
- **Better Spacing**: Increased gaps (1.5rem) between elements for reduced visual clutter
- **Responsive Grids**: Multi-column layouts that adapt to screen size using `auto-fit` and `minmax`
- **Enhanced Badges**: Larger status and role badges with better color contrast
- **Preserved Formatting**: Line breaks maintained in descriptions using `white-space: pre-wrap`
- **Color Coding**: Strategic use of primary, secondary, and text colors to show information importance
- **Clear Section Headers**: Semantic `<h2>` tags with consistent styling across all cards
- **Helpful Context**: Placeholder text and explanatory hints throughout forms

These improvements significantly enhance readability and make the interface easier to scan and navigate.

## User Interface

The application provides the following pages:

### For All Users
- **Login Page** - Email/password authentication
- **Force Password Change** - Required for newly created users on first login
- **Tickets List** - View all tickets with title, status badges, size, assignee, epic (shown as acronym), and sprint
  - Sorted by priority (promoted tickets at top), then by creation date (oldest first)
  - New unpromoted tickets appear at bottom of list
  - "Promote to Top" button to bump tickets to top priority
  - Click "View" to access ticket details, edit, and delete actions
- **Ticket Detail** - Enhanced card-based layout displaying ticket information in organized sections
  - Key Information card: Status, size, and assignee with email
  - Description card: Full description with preserved line breaks
  - Project Details card: Epic and sprint assignments with links
  - Metadata card: Creation and last updated timestamps
  - Edit and delete buttons enabled only for ticket creator or admin
- **Create/Edit Ticket** - Structured form with organized card sections
  - Basic Information card: Title and description with helpful placeholders
  - Assignment & Sizing card: Assignee and size in responsive 2-column grid
  - Project Organization card (edit only): Status, epic, and sprint assignment
  - Create mode: Required title and description, optional size and assignee (status defaults to "open")
  - Edit mode: Additional fields for status (5 options), epic, and sprint
- **Epics List** - View all epics with name column
  - Click "View" to access epic details, edit, and delete actions
- **Epic Detail** - Enhanced card-based layout for epic information
  - Epic Details card: Acronym (displayed prominently) and description with preserved line breaks
  - Tickets card: Table showing all tickets assigned to the epic with count in header
  - Edit and delete buttons available in detail view
- **Create/Edit Epic** - Structured form with organized sections
  - Epic Information card: Name and description with placeholders and acronym generation guidance
  - Helpful hint about using title case for clarity
- **Sprints List** - View all sprints with name, start date, and end date columns
  - "Board" and "View" buttons for accessing sprint board and details
- **Sprint Detail** - Enhanced card-based layout with intelligent status calculation
  - Sprint Details card: Active/Completed/Upcoming status badge, duration, and days remaining/until start
  - Timeline card: Start and end dates in responsive grid
  - Tickets card: Table showing all tickets assigned to the sprint with count in header
  - "View Board" button to access the interactive kanban board
  - Edit and delete buttons available in detail view
- **Sprint Board** - Interactive kanban board for sprint management
  - Three columns: Committed, Underway, Completed
  - Drag-and-drop tickets between columns to update status
  - Clickable status badges in Underway column (click to change between in-progress/blocked/needs-review)
  - Ticket cards show: ID, title, description (truncated), status badge, and view link
  - Ticket counts displayed in each column header
  - "Back to Details" button to return to sprint detail view
- **Create/Edit Sprint** - Structured form with organized sections
  - Sprint Information card: Name and start date with helpful placeholders
  - Clear explanation that end date is automatically set to 2 weeks after start date
- **Settings** - Enhanced card-based layout for account management
  - Account Information card: Full name, email, and role badge in responsive grid
  - Security card: Password change form with helpful security guidance

### Admin Only
- **User Management** - List all users with view access
  - Click "View" to access user details, edit, and delete actions
- **User Detail** - Enhanced card-based layout for user information
  - User Information card: Full name and email in responsive grid
  - Access & Permissions card: Role and account status badges
  - Edit and delete buttons available
- **Create/Edit User** - Structured form with organized sections
  - Personal Information card: First and last name in 2-column grid, email below
  - Security card (create only): Password field with hint about required change on first login
  - Access & Permissions card: Role selector with permission explanation, account status toggle (edit only)
  - New users must change password on first login
- **Delete Users** - Delete button available in user detail view
- **Delete Any Ticket/Epic/Sprint** - Delete button enabled for all tickets, epics, and sprints in their respective detail views

### Navigation
- Click the **Flyhalf** logo to return to the tickets list
- Click your **username** in the navbar to access settings
- **Tickets** link shows all tickets
- **Epics** link shows all epics
- **Sprints** link shows all sprints
  - From sprint detail page, click "View Board" to access the interactive kanban board
- **Users** link (admins only) for user management
- **Logout** button to end session
- **Active link highlighting** - Navbar automatically highlights the current section, including when viewing detail pages (e.g., viewing a specific ticket highlights the Tickets link)
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
- `description` (string, required)
- `status` (string: open, in-progress, blocked, needs-review, closed, default: 'open')
- `assigned_to` (UUID, optional) - Assign ticket to a user
- `epic_id` (UUID, optional) - Assign ticket to an epic
- `sprint_id` (UUID, optional) - Assign ticket to a sprint
- `size` (integer, optional: 1=Small, 2=Medium, 3=Large, 5=X-Large, 8=Danger)
- `priority` (integer, default: 0) - Automatically managed by promote feature

**Sprint Board Status Mapping**:
- Tickets with status `open` appear in the **Committed** column
- Tickets with status `in-progress`, `blocked`, or `needs-review` appear in the **Underway** column
- Tickets with status `closed` appear in the **Completed** column
- Dragging a ticket to a column updates its status accordingly (Committed→open, Underway→in-progress, Completed→closed)

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
- `description` (string, required)

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

### User Endpoints

| Method | Endpoint | Description | Auth Required | Role |
|--------|----------|-------------|---------------|------|
| GET | `/users` | List users for assignment | Yes | Any |

**Note**: Returns simplified user information (ID, name, email) for ticket assignment purposes. Available to all authenticated users.

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
- `description` (text, not null)
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
- `description` (text, not null)
- `status` (varchar(50): open, in-progress, blocked, needs-review, closed, default: 'open')
- `assigned_to` (UUID, FK to users, nullable)
- `epic_id` (UUID, FK to epics, nullable)
- `sprint_id` (UUID, FK to sprints, nullable)
- `size` (integer, nullable: 1=Small, 2=Medium, 3=Large, 5=X-Large, 8=Danger)
- `priority` (integer, default: 0)
- `created_at`, `updated_at` (timestamps)

## Security

- Passwords hashed with bcrypt (cost 12)
- Access tokens: Short-lived (15 minutes), stored in memory only
- Refresh tokens: HttpOnly + Secure + SameSite=Strict cookies
- CORS configured with explicit origin allowlist
- Parameterized queries to prevent SQL injection
- Security headers (X-Content-Type-Options, X-Frame-Options, etc.)

## Accessibility

- **Section 508 Compliant**: All color combinations meet WCAG 2.0 Level AA standards
- **Contrast Ratios**: Minimum 4.5:1 contrast ratio for normal text
- **Readable Colors**: Text and background colors optimized for readability
- **Status Indicators**: Color-coded status badges with sufficient contrast for visual accessibility
- **Interactive Elements**: Clear visual indicators for clickable and interactive elements

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

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

This means you are free to use, modify, and distribute this software, but any derivative works must also be licensed under GPL-3.0.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.
