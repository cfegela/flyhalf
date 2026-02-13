import { describe, it, expect } from 'vitest';
import {
    TICKET_STATUSES,
    STATUS_BADGE_CLASSES,
    TICKET_SIZES,
    SIZE_LABELS,
    CONSTRAINTS,
    UI_CONSTANTS
} from './tickets.js';

describe('TICKET_STATUSES', () => {
    it('should have all required statuses', () => {
        expect(TICKET_STATUSES.OPEN).toBe('open');
        expect(TICKET_STATUSES.IN_PROGRESS).toBe('in-progress');
        expect(TICKET_STATUSES.BLOCKED).toBe('blocked');
        expect(TICKET_STATUSES.NEEDS_REVIEW).toBe('needs-review');
        expect(TICKET_STATUSES.CLOSED).toBe('closed');
    });

    it('should have exactly 5 statuses', () => {
        const statusCount = Object.keys(TICKET_STATUSES).length;
        expect(statusCount).toBe(5);
    });
});

describe('STATUS_BADGE_CLASSES', () => {
    it('should have badge classes for all statuses', () => {
        expect(STATUS_BADGE_CLASSES['open']).toBe('badge-open');
        expect(STATUS_BADGE_CLASSES['in-progress']).toBe('badge-in-progress');
        expect(STATUS_BADGE_CLASSES['blocked']).toBe('badge-blocked');
        expect(STATUS_BADGE_CLASSES['needs-review']).toBe('badge-needs-review');
        expect(STATUS_BADGE_CLASSES['closed']).toBe('badge-closed');
    });

    it('should have exactly 5 badge classes', () => {
        const classCount = Object.keys(STATUS_BADGE_CLASSES).length;
        expect(classCount).toBe(5);
    });

    it('should have matching keys with TICKET_STATUSES values', () => {
        for (const statusValue of Object.values(TICKET_STATUSES)) {
            expect(STATUS_BADGE_CLASSES[statusValue]).toBeDefined();
        }
    });
});

describe('TICKET_SIZES', () => {
    it('should have all required sizes', () => {
        expect(TICKET_SIZES.SMALL).toBe(1);
        expect(TICKET_SIZES.MEDIUM).toBe(2);
        expect(TICKET_SIZES.LARGE).toBe(3);
        expect(TICKET_SIZES.X_LARGE).toBe(5);
        expect(TICKET_SIZES.DANGER).toBe(8);
    });

    it('should have exactly 5 sizes', () => {
        const sizeCount = Object.keys(TICKET_SIZES).length;
        expect(sizeCount).toBe(5);
    });

    it('should have sizes as numbers', () => {
        for (const size of Object.values(TICKET_SIZES)) {
            expect(typeof size).toBe('number');
        }
    });
});

describe('SIZE_LABELS', () => {
    it('should have labels for all sizes', () => {
        expect(SIZE_LABELS[1]).toBe('Small');
        expect(SIZE_LABELS[2]).toBe('Medium');
        expect(SIZE_LABELS[3]).toBe('Large');
        expect(SIZE_LABELS[5]).toBe('X-Large');
        expect(SIZE_LABELS[8]).toBe('Danger');
    });

    it('should have exactly 5 labels', () => {
        const labelCount = Object.keys(SIZE_LABELS).length;
        expect(labelCount).toBe(5);
    });

    it('should have matching keys with TICKET_SIZES values', () => {
        for (const sizeValue of Object.values(TICKET_SIZES)) {
            expect(SIZE_LABELS[sizeValue]).toBeDefined();
        }
    });

    it('should have labels as strings', () => {
        for (const label of Object.values(SIZE_LABELS)) {
            expect(typeof label).toBe('string');
        }
    });
});

describe('CONSTRAINTS', () => {
    it('should have UPDATE_MAX_LENGTH of 500', () => {
        expect(CONSTRAINTS.UPDATE_MAX_LENGTH).toBe(500);
    });

    it('should have CRITERIA_MAX_LENGTH of 256', () => {
        expect(CONSTRAINTS.CRITERIA_MAX_LENGTH).toBe(256);
    });

    it('should have MAX_CRITERIA_COUNT of 6', () => {
        expect(CONSTRAINTS.MAX_CRITERIA_COUNT).toBe(6);
    });

    it('should have exactly 3 constraints', () => {
        const constraintCount = Object.keys(CONSTRAINTS).length;
        expect(constraintCount).toBe(3);
    });

    it('should have all constraints as numbers', () => {
        for (const value of Object.values(CONSTRAINTS)) {
            expect(typeof value).toBe('number');
        }
    });
});

describe('UI_CONSTANTS', () => {
    it('should have DROPDOWN_OFFSET_PX of 5', () => {
        expect(UI_CONSTANTS.DROPDOWN_OFFSET_PX).toBe(5);
    });

    it('should have exactly 1 UI constant', () => {
        const constantCount = Object.keys(UI_CONSTANTS).length;
        expect(constantCount).toBe(1);
    });
});
