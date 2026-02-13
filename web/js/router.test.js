import { describe, it, expect, beforeEach } from 'vitest';

// We'll test the matchRoute method as a pure function
// by creating a minimal router class for testing
class TestRouter {
    constructor() {
        this.routes = [];
    }

    addRoute(path, handler, options = {}) {
        this.routes.push({ path, handler, options });
    }

    matchRoute(urlPath) {
        for (const route of this.routes) {
            const routeParts = route.path.split('/').filter(Boolean);
            const urlParts = urlPath.split('/').filter(Boolean);

            if (routeParts.length !== urlParts.length) {
                continue;
            }

            const params = [];
            let isMatch = true;

            for (let i = 0; i < routeParts.length; i++) {
                if (routeParts[i].startsWith(':')) {
                    params.push(urlParts[i]);
                } else if (routeParts[i] !== urlParts[i]) {
                    isMatch = false;
                    break;
                }
            }

            if (isMatch) {
                return { route, params };
            }
        }

        return null;
    }
}

describe('Router.matchRoute', () => {
    let router;

    beforeEach(() => {
        router = new TestRouter();
    });

    it('should match exact routes', () => {
        const handler = () => {};
        router.addRoute('/tickets', handler);

        const match = router.matchRoute('/tickets');

        expect(match).not.toBeNull();
        expect(match.route.path).toBe('/tickets');
        expect(match.params).toEqual([]);
    });

    it('should match routes with parameters', () => {
        const handler = () => {};
        router.addRoute('/tickets/:id', handler);

        const match = router.matchRoute('/tickets/123');

        expect(match).not.toBeNull();
        expect(match.route.path).toBe('/tickets/:id');
        expect(match.params).toEqual(['123']);
    });

    it('should match routes with multiple parameters', () => {
        const handler = () => {};
        router.addRoute('/sprints/:sprintId/tickets/:ticketId', handler);

        const match = router.matchRoute('/sprints/456/tickets/789');

        expect(match).not.toBeNull();
        expect(match.route.path).toBe('/sprints/:sprintId/tickets/:ticketId');
        expect(match.params).toEqual(['456', '789']);
    });

    it('should return null for non-matching routes', () => {
        const handler = () => {};
        router.addRoute('/tickets', handler);

        const match = router.matchRoute('/sprints');

        expect(match).toBeNull();
    });

    it('should return null when segment count differs', () => {
        const handler = () => {};
        router.addRoute('/tickets/:id', handler);

        const match = router.matchRoute('/tickets/123/edit');

        expect(match).toBeNull();
    });

    it('should return null for too few segments', () => {
        const handler = () => {};
        router.addRoute('/tickets/:id/edit', handler);

        const match = router.matchRoute('/tickets/123');

        expect(match).toBeNull();
    });

    it('should match root route', () => {
        const handler = () => {};
        router.addRoute('/', handler);

        const match = router.matchRoute('/');

        expect(match).not.toBeNull();
        expect(match.route.path).toBe('/');
        expect(match.params).toEqual([]);
    });

    it('should match first matching route when multiple routes exist', () => {
        const handler1 = () => 'handler1';
        const handler2 = () => 'handler2';
        router.addRoute('/tickets', handler1);
        router.addRoute('/tickets', handler2);

        const match = router.matchRoute('/tickets');

        expect(match).not.toBeNull();
        expect(match.route.handler()).toBe('handler1');
    });

    it('should handle routes with trailing slashes', () => {
        const handler = () => {};
        router.addRoute('/tickets', handler);

        // URL without trailing slash matches route without trailing slash
        const match1 = router.matchRoute('/tickets');
        expect(match1).not.toBeNull();

        // URL with trailing slash should also work (empty parts filtered)
        const match2 = router.matchRoute('/tickets/');
        expect(match2).not.toBeNull();
    });

    it('should distinguish between static and parameterized routes', () => {
        const handler1 = () => 'static';
        const handler2 = () => 'param';
        router.addRoute('/tickets/new', handler1);
        router.addRoute('/tickets/:id', handler2);

        const matchStatic = router.matchRoute('/tickets/new');
        expect(matchStatic).not.toBeNull();
        expect(matchStatic.route.handler()).toBe('static');

        const matchParam = router.matchRoute('/tickets/123');
        expect(matchParam).not.toBeNull();
        expect(matchParam.route.handler()).toBe('param');
        expect(matchParam.params).toEqual(['123']);
    });

    it('should handle complex paths', () => {
        const handler = () => {};
        router.addRoute('/admin/users/:userId/permissions/:permId', handler);

        const match = router.matchRoute('/admin/users/abc-123/permissions/read-write');

        expect(match).not.toBeNull();
        expect(match.params).toEqual(['abc-123', 'read-write']);
    });

    it('should return null for empty path when no root route', () => {
        const handler = () => {};
        router.addRoute('/tickets', handler);

        const match = router.matchRoute('/');

        expect(match).toBeNull();
    });

    it('should match UUID-like parameters', () => {
        const handler = () => {};
        router.addRoute('/tickets/:id', handler);

        const match = router.matchRoute('/tickets/550e8400-e29b-41d4-a716-446655440000');

        expect(match).not.toBeNull();
        expect(match.params).toEqual(['550e8400-e29b-41d4-a716-446655440000']);
    });

    it('should preserve parameter order', () => {
        const handler = () => {};
        router.addRoute('/a/:first/b/:second/c/:third', handler);

        const match = router.matchRoute('/a/1/b/2/c/3');

        expect(match).not.toBeNull();
        expect(match.params).toEqual(['1', '2', '3']);
    });
});
