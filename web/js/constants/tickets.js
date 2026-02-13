// Ticket status values
export const TICKET_STATUSES = {
    OPEN: 'open',
    IN_PROGRESS: 'in-progress',
    BLOCKED: 'blocked',
    NEEDS_REVIEW: 'needs-review',
    CLOSED: 'closed'
};

// Status badge CSS class mapping
export const STATUS_BADGE_CLASSES = {
    'open': 'badge-open',
    'in-progress': 'badge-in-progress',
    'blocked': 'badge-blocked',
    'needs-review': 'badge-needs-review',
    'closed': 'badge-closed'
};

// Ticket size values
export const TICKET_SIZES = {
    SMALL: 1,
    MEDIUM: 2,
    LARGE: 3,
    X_LARGE: 5,
    DANGER: 8
};

// Size display labels
export const SIZE_LABELS = {
    1: 'Small',
    2: 'Medium',
    3: 'Large',
    5: 'X-Large',
    8: 'Danger'
};

// Validation constraints
export const CONSTRAINTS = {
    UPDATE_MAX_LENGTH: 500,
    CRITERIA_MAX_LENGTH: 256,
    MAX_CRITERIA_COUNT: 6
};

// UI constants
export const UI_CONSTANTS = {
    DROPDOWN_OFFSET_PX: 5
};
