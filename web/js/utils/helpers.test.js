import { describe, it, expect } from 'vitest';
import { createIdMap, getStatusBadgeClass, getSizeLabel, getProjectAcronym } from './helpers.js';

describe('createIdMap', () => {
    it('should create a map from an array of objects', () => {
        const items = [
            { id: '1', name: 'Item 1' },
            { id: '2', name: 'Item 2' },
            { id: '3', name: 'Item 3' }
        ];

        const map = createIdMap(items);

        expect(map).toEqual({
            '1': { id: '1', name: 'Item 1' },
            '2': { id: '2', name: 'Item 2' },
            '3': { id: '3', name: 'Item 3' }
        });
    });

    it('should handle empty array', () => {
        const items = [];
        const map = createIdMap(items);
        expect(map).toEqual({});
    });

    it('should handle single item', () => {
        const items = [{ id: 'abc', value: 'test' }];
        const map = createIdMap(items);
        expect(map).toEqual({
            'abc': { id: 'abc', value: 'test' }
        });
    });
});

describe('getStatusBadgeClass', () => {
    it('should return correct class for open status', () => {
        expect(getStatusBadgeClass('open')).toBe('badge-open');
    });

    it('should return correct class for in-progress status', () => {
        expect(getStatusBadgeClass('in-progress')).toBe('badge-in-progress');
    });

    it('should return correct class for blocked status', () => {
        expect(getStatusBadgeClass('blocked')).toBe('badge-blocked');
    });

    it('should return correct class for needs-review status', () => {
        expect(getStatusBadgeClass('needs-review')).toBe('badge-needs-review');
    });

    it('should return correct class for closed status', () => {
        expect(getStatusBadgeClass('closed')).toBe('badge-closed');
    });

    it('should return default class for unknown status', () => {
        expect(getStatusBadgeClass('unknown')).toBe('badge-open');
    });

    it('should return default class for null status', () => {
        expect(getStatusBadgeClass(null)).toBe('badge-open');
    });

    it('should return default class for undefined status', () => {
        expect(getStatusBadgeClass(undefined)).toBe('badge-open');
    });
});

describe('getSizeLabel', () => {
    it('should return "Small" for size 1', () => {
        expect(getSizeLabel(1)).toBe('Small');
    });

    it('should return "Medium" for size 2', () => {
        expect(getSizeLabel(2)).toBe('Medium');
    });

    it('should return "Large" for size 3', () => {
        expect(getSizeLabel(3)).toBe('Large');
    });

    it('should return "X-Large" for size 5', () => {
        expect(getSizeLabel(5)).toBe('X-Large');
    });

    it('should return "Danger" for size 8', () => {
        expect(getSizeLabel(8)).toBe('Danger');
    });

    it('should return "-" for unknown size', () => {
        expect(getSizeLabel(99)).toBe('-');
    });

    it('should return "-" for null size', () => {
        expect(getSizeLabel(null)).toBe('-');
    });

    it('should return "-" for undefined size', () => {
        expect(getSizeLabel(undefined)).toBe('-');
    });

    it('should return "-" for zero size', () => {
        expect(getSizeLabel(0)).toBe('-');
    });
});

describe('getProjectAcronym', () => {
    it('should return first 6 characters with spaces removed', () => {
        const projects = [{ id: '1', name: 'Project One' }];
        const result = getProjectAcronym('Project One', projects, '1');
        expect(result).toBe('PROJEC');
    });

    it('should return first 6 characters when multiple projects exist', () => {
        const projects = [
            { id: '1', name: 'Project One' },
            { id: '2', name: 'Another Thing' }
        ];
        const result = getProjectAcronym('Project One', projects, '1');
        expect(result).toBe('PROJEC');
    });

    it('should return first 6 characters for different project', () => {
        const projects = [
            { id: '1', name: 'Project One' },
            { id: '2', name: 'Another Thing' }
        ];
        const result = getProjectAcronym('Another Thing', projects, '2');
        expect(result).toBe('ANOTHE');
    });

    it('should return first 6 characters for multi-word projects', () => {
        const projects = [
            { id: '1', name: 'Hello World' },
            { id: '2', name: 'Foo Bar' }
        ];
        const result = getProjectAcronym('Hello World', projects, '1');
        expect(result).toBe('HELLOW');
    });

    it('should return all characters if less than 6 characters', () => {
        const projects = [
            { id: '1', name: 'Alpha' },
            { id: '2', name: 'Beta' }
        ];
        const result = getProjectAcronym('Alpha', projects, '1');
        expect(result).toBe('ALPHA');
    });

    it('should return all characters for very short project names', () => {
        const projects = [
            { id: '1', name: 'Ab' },
            { id: '2', name: 'Beta' }
        ];
        const result = getProjectAcronym('Ab', projects, '1');
        expect(result).toBe('AB');
    });

    it('should return first 6 characters for long multi-word names', () => {
        const projects = [
            { id: '1', name: 'One Two Three Four' },
            { id: '2', name: 'Beta' }
        ];
        const result = getProjectAcronym('One Two Three Four', projects, '1');
        expect(result).toBe('ONETWO');
    });

    it('should extend beyond 6 characters when collision detected', () => {
        const projects = [
            { id: '1', name: 'Demo Project' },
            { id: '2', name: 'Demo Proposal' }
        ];
        // "Demo Project" -> "DemoProject" (11 chars)
        // "Demo Proposal" -> "DemoProposal" (12 chars)
        // At length 6: both "DEMOPRO" (collision!)
        // At length 7: "DEMOPROJ" vs "DEMOPROP" (no collision)
        // At length 8: "DEMOPROJ" vs "DEMOPROP" (no collision)
        const result1 = getProjectAcronym('Demo Project', projects, '1');
        const result2 = getProjectAcronym('Demo Proposal', projects, '2');
        expect(result1).toBe('DEMOPROJ');
        expect(result2).toBe('DEMOPROP');
    });

    it('should return "-" for null project name', () => {
        const projects = [{ id: '1', name: 'Project' }];
        const result = getProjectAcronym(null, projects, null);
        expect(result).toBe('-');
    });

    it('should return "-" for undefined project name', () => {
        const projects = [{ id: '1', name: 'Project' }];
        const result = getProjectAcronym(undefined, projects, null);
        expect(result).toBe('-');
    });

    it('should return "-" for empty project name', () => {
        const projects = [{ id: '1', name: 'Project' }];
        const result = getProjectAcronym('', projects, null);
        expect(result).toBe('-');
    });
});
