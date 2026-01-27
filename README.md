# Flyhalf

### A Very, Very Opinionated Agile Scrum Management System

**flyhalf /ˈflaɪˌhɑːf/ noun**

The flyhalf is rugby’s primary playmaker and tactical leader who directs the team's attack. Flyhalves possess exceptional vision, game management, accuracy, and resilience.

## Tech Stack

- **Backend**: Go 1.24 with chi router, pgx (PostgreSQL driver), golang-jwt
- **Frontend**: Vanilla JavaScript SPA with ES modules (no build step required)
- **Database**: PostgreSQL 16
- **Authentication**: JWT access tokens (15min) + refresh tokens (7 days, HttpOnly cookie)
- **Development**: Docker Compose or Podman Compose with hot reload (Air for Go)

## Features

- JWT-based authentication with token refresh
- User-friendly login error messages for common issues (invalid credentials, inactive accounts, network errors)
- Role-based access control (admin/user)
- **Ticket Management**:
  - CRUD operations with 5 status options (open, in-progress, blocked, needs-review, closed)
  - New tickets automatically default to "open" status
  - New tickets appear at bottom of list (priority 0, promoted tickets at top)
  - **Flexible Priority Management** with icon-based action buttons (using Heroicons):
    - Promote to top: Send ticket to highest priority
    - Promote up one: Move ticket up one position
    - Demote down one: Move ticket down one position
    - Priorities persist between application restarts
  - Optional ticket sizing (Small=1, Medium=2, Large=3, X-Large=5, Danger=8)
  - Ticket assignment to users with assignee display in ticket list
  - Assign tickets to projects for organization (project names shown as acronyms in list)
  - Assign tickets to sprints for sprint planning
  - Required title and description fields
  - 6-character unique ID for each ticket
  - Quick actions in list view: View and Edit buttons with clean icon interface
- **Project Management**:
  - CRUD operations for projects with required name and description fields
  - Organize tickets by assigning them to projects
  - Project detail view shows all tickets assigned to that project
  - Project names displayed as acronyms in ticket list (first 6 characters excluding spaces, uppercased)
  - Full list and detail views
  - Simplified list view with detail-level actions (edit/delete available in detail view only)
- **Sprint Management**:
  - CRUD operations for sprints (name and start date)
  - End date automatically calculated as 2 weeks after start date
  - **Dynamic Status Calculation**: Sprint status (upcoming/active/completed) automatically calculated based on start/end dates
  - Assign tickets to sprints for sprint planning
  - Sprint detail view shows all tickets assigned to that sprint
  - **Sprint Board**: Interactive kanban board with drag-and-drop functionality
    - Three columns: Committed (open), Underway (in-progress/blocked/needs-review), Completed (closed)
    - Drag tickets between columns to update their status
    - Clickable status badges in Underway column to change between in-progress/blocked/needs-review
    - Real-time status updates via API
    - Tickets sorted by priority within each column
    - Responsive design with mobile support
  - **Sprint Report**: Visual burndown chart and progress metrics
    - Story points and ticket completion metrics with progress bars
    - Burndown chart showing remaining story points over sprint duration
    - Breakdown by total/completed/remaining points and tickets
    - Powered by Chart.js for interactive visualization
  - Full list and detail views
  - Simplified list view with Board and Report buttons for quick access
- All users can view and edit all tickets and projects (collaborative workspace)
- Users can delete tickets/projects they created; admins can delete any ticket/project
- Forced password change for newly created users
- Admin user management
- User settings page with account information
- Password change functionality
- Responsive UI with modern CSS
- Client-side routing with hash-based navigation that preserves state on refresh
- Secure HttpOnly cookies for refresh tokens

## Permission Model

### Regular Users (role: 'user')
- ✅ View all tickets, projects, and sprints
- ✅ Create new tickets, projects, and sprints
- ✅ Edit any ticket, project, or sprint
- ✅ Delete tickets, projects, and sprints they created
- ✅ Assign tickets to projects and sprints
- ✅ Manage ticket priorities (promote to top, promote up, demote down)
- ✅ Use sprint board with drag-and-drop to update ticket status
- ✅ Change own password
- ✅ View own account settings
- ❌ Delete tickets/projects/sprints created by others
- ❌ Manage users

### Administrators (role: 'admin')
- ✅ All user permissions
- ✅ Delete any ticket, project, or sprint (including those created by others)
- ✅ Create new users (with forced password change)
- ✅ Edit user accounts
- ✅ Delete users
- ✅ Deactivate/activate users

This collaborative permission model allows all team members to view and update tickets, projects, and sprints while protecting data integrity. Users can manage their own items completely, but cannot delete items created by others.

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

- Docker and Docker Compose (or Podman and Podman Compose)
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
# or with Podman
podman compose up
```

This will start three services:
- **PostgreSQL** on port 5432
- **API** on port 8080
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

### Latest Updates (January 2026)
- **Dynamic Sprint Status Calculation**: Sprint status (upcoming/active/completed) now calculated server-side based on current date and sprint timeline, ensuring consistent status across all views
- **Improved Date Handling**: All date parsing now uses local timezone to prevent timezone conversion issues in sprint boards, reports, and detail views
- **Icon-Based Ticket Priority Controls**: Tickets list now features a modern actions column with 5 Heroicons SVG icons:
  - Promote to top (⇈): Send ticket to highest priority
  - Promote up one (↑): Swap priority with ticket immediately above
  - Demote down one (↓): Swap priority with ticket immediately below
  - View (👁): Navigate to ticket details
  - Edit (✏️): Navigate to ticket edit form
  - Clean, accessible icons with proper tooltips and responsive sizing
- **Enhanced Login Error Handling**: User-friendly error messages now appear on the login page for common failure scenarios:
  - Invalid credentials with clear guidance
  - Inactive account notifications
  - Network/connection error messages
  - Styled error alerts with borders and background color for better visibility
- **Fixed Ticket Priority Persistence**: Ticket priorities now correctly persist between application restarts (previously would reset to 0)
- **API Port Standardization**: API now runs on port 8080 for both host and container (previously mapped 8081:8080)

### Structured Card Layouts
All detail pages and forms now use organized card-based layouts with clear section headers:

**Detail Pages**:
- **Ticket Details**: Organized into Key Information, Description, Project Details, and Metadata cards
- **Project Details**: Organized into Project Details (with acronym display) and Tickets cards
- **Sprint Details**: Organized into Sprint Details (with status calculation), Timeline, and Tickets cards
  - Shows active/completed/upcoming status with color-coded badges
  - Displays duration and days remaining/until start
- **User Details**: Organized into User Information and Access & Permissions cards
- **Settings Page**: Organized into Account Information and Security cards

**Create/Edit Forms**:
- **Ticket Form**: Organized into Basic Information, Assignment & Sizing, Project Organization (edit only), and Form Actions cards
  - Responsive 2-column grids for related fields
  - Helpful placeholders and contextual hints
- **Project Form**: Organized into Project Information and Form Actions cards
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
- **Tickets List** - View all tickets with title, status badges, size, assignee, project (shown as acronym), and sprint
  - Sorted by priority (promoted tickets at top), then by creation date (oldest first)
  - New unpromoted tickets appear at bottom of list
  - **Icon-based Actions Column** with 5 intuitive controls:
    - ⇈ Promote to top: Send ticket to highest priority
    - ↑ Promote up one: Swap priority with ticket above
    - ↓ Demote down one: Swap priority with ticket below
    - 👁 View: Access ticket details
    - ✏️ Edit: Modify ticket information
  - Clean Heroicons SVG icons for modern, accessible interface
- **Ticket Detail** - Enhanced card-based layout displaying ticket information in organized sections
  - Key Information card: Status, size, and assignee with email
  - Description card: Full description with preserved line breaks
  - Project Details card: Project and sprint assignments with links
  - Metadata card: Creation and last updated timestamps
  - Edit and delete buttons enabled only for ticket creator or admin
- **Create/Edit Ticket** - Structured form with organized card sections
  - Basic Information card: Title and description with helpful placeholders
  - Assignment & Sizing card: Assignee and size in responsive 2-column grid
  - Project Organization card (edit only): Status, project, and sprint assignment
  - Create mode: Required title and description, optional size and assignee (status defaults to "open")
  - Edit mode: Additional fields for status (5 options), project, and sprint
- **Projects List** - View all projects with name column
  - Click "View" to access project details, edit, and delete actions
- **Project Detail** - Enhanced card-based layout for project information
  - Project Details card: Acronym (first 6 characters excluding spaces, uppercased) and description with preserved line breaks
  - Tickets card: Table showing all tickets assigned to the project with count in header
  - Edit and delete buttons available in detail view
- **Create/Edit Project** - Structured form with organized sections
  - Project Information card: Name and description with placeholders and acronym generation guidance
  - Helpful hint explaining first 6 characters (excluding spaces) form the acronym
- **Sprints List** - View all sprints with name, status, start date, and end date columns
  - "Board" and "Report" buttons for quick access to sprint board and analytics
  - Click sprint row to view full details
- **Sprint Detail** - Enhanced card-based layout with intelligent status calculation
  - Sprint Details card: Active/Completed/Upcoming status badge, duration, and days remaining/until start
  - Timeline card: Start and end dates in responsive grid
  - Tickets card: Table showing all tickets assigned to the sprint with count in header
  - "View Board" and "View Report" buttons to access the interactive kanban board and analytics
  - Edit and delete buttons available in detail view
- **Sprint Board** - Interactive kanban board for sprint management
  - Three columns: Committed, Underway, Completed
  - Drag-and-drop tickets between columns to update status
  - Clickable status badges in Underway column (click to change between in-progress/blocked/needs-review)
  - Ticket cards show: ID, title, description (truncated), status badge, and view link
  - Ticket counts displayed in each column header
  - "Back to Details" button to return to sprint detail view
- **Sprint Report** - Visual analytics and burndown tracking
  - Story Points card: Total, completed, and remaining points with progress bar and percentage
  - Tickets card: Total and completed ticket count with progress bar and percentage
  - Burndown Chart: Line graph showing remaining story points across sprint duration
  - Chart powered by Chart.js with interactive tooltips
  - Clean, focused visualization for sprint progress tracking
  - "View Board" and "Back to Details" buttons for easy navigation
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
- **Delete Any Ticket/Project/Sprint** - Delete button enabled for all tickets, projects, and sprints in their respective detail views

### Navigation
- Click the **Flyhalf** logo to return to the tickets list
- Click your **username** in the navbar to access settings
- **Tickets** link shows all tickets
- **Projects** link shows all projects
- **Sprints** link shows all sprints
  - From sprints list, click "Board" or "Report" buttons for quick access
  - From sprint detail page, click "View Board" for the interactive kanban board or "View Report" for analytics
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

**Local Development:**
```
http://localhost:8080/api/v1
```

**Production (GCP):**
```
https://flyhalf-prod-api-oas33witna-uc.a.run.app/api/v1
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
| POST | `/tickets/{id}/promote-up` | Promote ticket up one position | Yes | Any |
| POST | `/tickets/{id}/demote-down` | Demote ticket down one position | Yes | Any |

**Note**: All authenticated users can view and edit all tickets. Users can delete tickets they created; admins can delete any ticket.

**Ticket Fields**:
- `title` (string, required)
- `description` (string, required)
- `status` (string: open, in-progress, blocked, needs-review, closed, default: 'open')
- `assigned_to` (UUID, optional) - Assign ticket to a user
- `project_id` (UUID, optional) - Assign ticket to a project
- `sprint_id` (UUID, optional) - Assign ticket to a sprint
- `size` (integer, optional: 1=Small, 2=Medium, 3=Large, 5=X-Large, 8=Danger)
- `priority` (integer, default: 0) - Automatically managed by promote feature

**Sprint Board Status Mapping**:
- Tickets with status `open` appear in the **Committed** column
- Tickets with status `in-progress`, `blocked`, or `needs-review` appear in the **Underway** column
- Tickets with status `closed` appear in the **Completed** column
- Dragging a ticket to a column updates its status accordingly (Committed→open, Underway→in-progress, Completed→closed)

### Project Endpoints

| Method | Endpoint | Description | Auth Required | Role |
|--------|----------|-------------|---------------|------|
| GET | `/projects` | List all projects | Yes | Any |
| POST | `/projects` | Create project | Yes | Any |
| GET | `/projects/{id}` | Get project by ID | Yes | Any |
| PUT | `/projects/{id}` | Update project | Yes | Any |
| DELETE | `/projects/{id}` | Delete project | Yes | Creator or Admin |

**Note**: All authenticated users can view and edit all projects. Users can delete projects they created; admins can delete any project.

**Project Fields**:
- `name` (string, required)
- `description` (string, required)

### Sprint Endpoints

| Method | Endpoint | Description | Auth Required | Role |
|--------|----------|-------------|---------------|------|
| GET | `/sprints` | List all sprints | Yes | Any |
| POST | `/sprints` | Create sprint | Yes | Any |
| GET | `/sprints/{id}` | Get sprint by ID | Yes | Any |
| GET | `/sprints/{id}/report` | Get sprint report data | Yes | Any |
| PUT | `/sprints/{id}` | Update sprint | Yes | Any |
| DELETE | `/sprints/{id}` | Delete sprint | Yes | Creator or Admin |

**Note**: All authenticated users can view and edit all sprints. Users can delete sprints they created; admins can delete any sprint.

**Sprint Fields**:
- `name` (string, required)
- `start_date` (date, required, format: YYYY-MM-DD)
- `end_date` (date, auto-calculated as start_date + 14 days)

**Sprint Report Response**:
The `/sprints/{id}/report` endpoint returns comprehensive sprint analytics including:
- `sprint` - Full sprint object
- `total_points` - Total story points for all tickets in sprint
- `completed_points` - Story points for closed tickets
- `remaining_points` - Story points for open/in-progress tickets
- `total_tickets` - Count of all tickets in sprint
- `completed_tickets` - Count of closed tickets
- `ideal_burndown` - Array of daily burndown points (date, points)
- `tickets_by_status` - Map of ticket counts grouped by status
- `points_by_status` - Map of story points grouped by status

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

### Projects Table
- `id` (UUID, primary key)
- `user_id` (FK to users - project creator)
- `name` (varchar(255), not null)
- `description` (text, not null)
- `created_at`, `updated_at` (timestamps)

### Sprints Table
- `id` (UUID, primary key)
- `user_id` (FK to users - sprint creator)
- `name` (varchar(255), not null)
- `start_date` (date, not null)
- `end_date` (date, not null)
- `status` (computed field: upcoming/active/completed, calculated based on current date)
- `created_at`, `updated_at` (timestamps)

### Tickets Table
- `id` (UUID, primary key)
- `user_id` (FK to users - ticket creator)
- `title` (varchar(255), not null)
- `description` (text, not null)
- `status` (varchar(50): open, in-progress, blocked, needs-review, closed, default: 'open')
- `assigned_to` (UUID, FK to users, nullable)
- `project_id` (UUID, FK to projects, nullable)
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

### Self-Hosted Deployment

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

### Google Cloud Platform (GCP) Deployment

The application is currently deployed to GCP using Cloud Run, Cloud SQL, and Cloud Storage with CDN.

**Live Application**: https://www.flyhalf.app

**Production Admin Credentials:**
- Email: `admin@flyhalf.local`
- Password: `admin123`
- **IMPORTANT**: Change this password immediately after first login!

#### Infrastructure Overview

- **Frontend**: Cloud Storage + Cloud CDN with global HTTPS load balancer
- **API**: Cloud Run (serverless containers) at https://flyhalf-prod-api-oas33witna-uc.a.run.app
- **Database**: Cloud SQL PostgreSQL 16 (private IP only)
- **Secrets**: Secret Manager for sensitive credentials (DB password, JWT secrets)
- **Region**: us-central1
- **SSL**: Google-managed certificates for HTTPS

#### Deployment Configuration

All GCP infrastructure is managed with Terraform in `ops/terraform/gcp/`. The configuration includes:

- **Networking**: Custom VPC with Serverless VPC Access Connector for Cloud Run → Cloud SQL
- **Security**: Service accounts with least privilege, no public database IP
- **Monitoring**: Health checks on `/health` endpoint
- **Scaling**: Min 0 instances (scales to zero), max 10 instances
- **Cost**: Estimated $15-25/month for production workload

#### CORS Configuration

The API requires the `ALLOWED_ORIGIN` environment variable to be set for frontend access:
```bash
ALLOWED_ORIGIN=https://www.flyhalf.app
```

Without proper CORS configuration, the frontend will display "Unable to connect to server" errors.

#### Database Migrations

Database migrations run automatically when the API starts. The default admin user is created during the initial migration.

#### Terraform Deployment

See `ops/terraform/gcp/README.md` for detailed Terraform deployment instructions including:
- Required GCP APIs and permissions
- Creating secrets in Secret Manager
- Building and pushing Docker images to Artifact Registry
- Uploading frontend files to Cloud Storage
- DNS configuration for custom domains

**Estimated Monthly Costs:**
- Cloud Run API: $5-10 (minimal traffic, scales to zero)
- Cloud SQL: $10-15 (db-f1-micro instance)
- Cloud Storage + CDN: $1-2
- Other (networking, secrets): $1-2
- **Total**: $15-25/month

## Troubleshooting

### Port Already in Use

If ports 3000, 5432, or 8080 are already in use, modify the port mappings in `docker-compose.yml`.

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
