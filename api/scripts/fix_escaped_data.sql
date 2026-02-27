-- Fix double/multi-escaped HTML entities in tickets table
-- This script unescapes HTML entities that were incorrectly escaped on input

-- Create a temporary function to recursively unescape HTML entities
CREATE OR REPLACE FUNCTION unescape_html(text) RETURNS text AS $$
DECLARE
    result text := $1;
    prev_result text;
BEGIN
    -- Keep unescaping until no more changes occur (handles multiple levels of escaping)
    LOOP
        prev_result := result;
        
        -- Unescape common HTML entities
        result := replace(result, '&amp;', '&');
        result := replace(result, '&lt;', '<');
        result := replace(result, '&gt;', '>');
        result := replace(result, '&quot;', '"');
        result := replace(result, '&#39;', '''');
        result := replace(result, '&#x27;', '''');
        result := replace(result, '&apos;', '''');
        
        -- If no changes were made, we're done
        EXIT WHEN result = prev_result;
    END LOOP;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Update tickets table
UPDATE tickets
SET 
    title = unescape_html(title),
    description = unescape_html(description)
WHERE 
    title LIKE '%&amp;%' OR title LIKE '%&#%' OR
    description LIKE '%&amp;%' OR description LIKE '%&#%';

-- Update acceptance_criteria table
UPDATE acceptance_criteria
SET content = unescape_html(content)
WHERE content LIKE '%&amp;%' OR content LIKE '%&#%';

-- Update ticket_updates table
UPDATE ticket_updates
SET content = unescape_html(content)
WHERE content LIKE '%&amp;%' OR content LIKE '%&#%';

-- Update projects table
UPDATE projects
SET 
    name = unescape_html(name),
    description = unescape_html(description)
WHERE 
    name LIKE '%&amp;%' OR name LIKE '%&#%' OR
    description LIKE '%&amp;%' OR description LIKE '%&#%';

-- Clean up the temporary function
DROP FUNCTION unescape_html(text);

-- Show summary of affected rows
SELECT 'Migration complete - fixed escaped HTML entities' as status;
