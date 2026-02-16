import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { escapeHtml, formatDate, formatDateOnly, formatRelativeTime } from './formatting.js';

describe('escapeHtml', () => {
    it('should escape HTML special characters', () => {
        const input = '<script>alert("xss")</script>';
        const result = escapeHtml(input);
        expect(result).toBe('&lt;script&gt;alert("xss")&lt;/script&gt;');
    });

    it('should escape ampersands', () => {
        const input = 'A & B';
        const result = escapeHtml(input);
        expect(result).toBe('A &amp; B');
    });

    it('should escape quotes', () => {
        const input = '"quoted" text';
        const result = escapeHtml(input);
        expect(result).toBe('"quoted" text'); // textContent doesn't escape quotes
    });

    it('should handle plain text', () => {
        const input = 'Hello, World!';
        const result = escapeHtml(input);
        expect(result).toBe('Hello, World!');
    });

    it('should handle empty string', () => {
        const input = '';
        const result = escapeHtml(input);
        expect(result).toBe('');
    });

    it('should handle multiple special characters', () => {
        const input = '<div class="test">A & B</div>';
        const result = escapeHtml(input);
        expect(result).toBe('&lt;div class="test"&gt;A &amp; B&lt;/div&gt;');
    });
});

describe('formatDate', () => {
    it('should format a date with time', () => {
        const dateString = '2024-01-15T14:30:00Z';
        const result = formatDate(dateString);
        // Result will vary by locale, but should contain date and time
        expect(result).toBeTruthy();
        expect(result).toContain('2024'); // Year should be present
    });

    it('should handle different date formats', () => {
        const dateString = '2024-12-25T00:00:00Z';
        const result = formatDate(dateString);
        expect(result).toBeTruthy();
    });
});

describe('formatDateOnly', () => {
    it('should format a date without time', () => {
        const dateString = '2024-01-15T14:30:00Z';
        const result = formatDateOnly(dateString);
        // Result will vary by locale, but should contain date without time
        expect(result).toBeTruthy();
        expect(result).toContain('2024'); // Year should be present
        expect(result).not.toContain(':'); // Should not contain time separator
    });

    it('should handle date-only strings', () => {
        const dateString = '2024-12-25';
        const result = formatDateOnly(dateString);
        expect(result).toBeTruthy();
        expect(result).not.toContain(':'); // Should not contain time separator
    });

    it('should handle different date formats', () => {
        const dateString = '2024-12-25T00:00:00Z';
        const result = formatDateOnly(dateString);
        expect(result).toBeTruthy();
        expect(result).not.toContain(':'); // Should not contain time separator
    });
});

describe('formatRelativeTime', () => {
    beforeEach(() => {
        // Mock the current time to 2024-01-15 12:00:00
        vi.useFakeTimers();
        vi.setSystemTime(new Date('2024-01-15T12:00:00Z'));
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('should return "just now" for times less than 60 seconds ago', () => {
        const dateString = '2024-01-15T11:59:30Z'; // 30 seconds ago
        const result = formatRelativeTime(dateString);
        expect(result).toBe('just now');
    });

    it('should return minutes for times less than 60 minutes ago', () => {
        const dateString = '2024-01-15T11:55:00Z'; // 5 minutes ago
        const result = formatRelativeTime(dateString);
        expect(result).toBe('5 minutes ago');
    });

    it('should return singular minute', () => {
        const dateString = '2024-01-15T11:59:00Z'; // 1 minute ago
        const result = formatRelativeTime(dateString);
        expect(result).toBe('1 minute ago');
    });

    it('should return hours for times less than 24 hours ago', () => {
        const dateString = '2024-01-15T09:00:00Z'; // 3 hours ago
        const result = formatRelativeTime(dateString);
        expect(result).toBe('3 hours ago');
    });

    it('should return singular hour', () => {
        const dateString = '2024-01-15T11:00:00Z'; // 1 hour ago
        const result = formatRelativeTime(dateString);
        expect(result).toBe('1 hour ago');
    });

    it('should return days for times less than 7 days ago', () => {
        const dateString = '2024-01-13T12:00:00Z'; // 2 days ago
        const result = formatRelativeTime(dateString);
        expect(result).toBe('2 days ago');
    });

    it('should return singular day', () => {
        const dateString = '2024-01-14T12:00:00Z'; // 1 day ago
        const result = formatRelativeTime(dateString);
        expect(result).toBe('1 day ago');
    });

    it('should return formatted date for times 7+ days ago', () => {
        const dateString = '2024-01-01T12:00:00Z'; // 14 days ago
        const result = formatRelativeTime(dateString);
        // Should fall back to formatDate
        expect(result).toBeTruthy();
        expect(result).not.toContain('ago');
    });

    it('should handle edge case of exactly 60 seconds', () => {
        const dateString = '2024-01-15T11:59:00Z'; // 60 seconds ago
        const result = formatRelativeTime(dateString);
        expect(result).toBe('1 minute ago');
    });

    it('should handle edge case of exactly 60 minutes', () => {
        const dateString = '2024-01-15T11:00:00Z'; // 60 minutes ago
        const result = formatRelativeTime(dateString);
        expect(result).toBe('1 hour ago');
    });

    it('should handle edge case of exactly 24 hours', () => {
        const dateString = '2024-01-14T12:00:00Z'; // 24 hours ago
        const result = formatRelativeTime(dateString);
        expect(result).toBe('1 day ago');
    });

    it('should handle edge case of exactly 7 days', () => {
        const dateString = '2024-01-08T12:00:00Z'; // 7 days ago
        const result = formatRelativeTime(dateString);
        // Should fall back to formatDate
        expect(result).toBeTruthy();
        expect(result).not.toContain('ago');
    });
});
