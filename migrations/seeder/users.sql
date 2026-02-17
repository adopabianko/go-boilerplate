-- Seed users
-- Uses ON CONFLICT to avoid duplicate email errors

INSERT INTO users (email, password)
VALUES 
    ('seed1@example.com', '$2a$10$.fn/MHkPqrdNyhuk7f95/O/Lo10q7KsQJqhDnm0V5E7rzFmY8vhVq'), -- password: password123
    ('seed2@example.com', '$2a$10$.fn/MHkPqrdNyhuk7f95/O/Lo10q7KsQJqhDnm0V5E7rzFmY8vhVq'), -- password: password123
    ('seed3@example.com', '$2a$10$.fn/MHkPqrdNyhuk7f95/O/Lo10q7KsQJqhDnm0V5E7rzFmY8vhVq'), -- password: password123
    ('seed4@example.com', '$2a$10$.fn/MHkPqrdNyhuk7f95/O/Lo10q7KsQJqhDnm0V5E7rzFmY8vhVq')    -- password: password123
ON CONFLICT (email) DO NOTHING;
