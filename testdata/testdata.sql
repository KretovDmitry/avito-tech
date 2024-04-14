INSERT INTO banners (feature_id, title, text, url, is_active, is_deleted)
    VALUES (1, 'some_title', 'some_text', 'some_url', TRUE, FALSE),
    (2, 'some_title', 'some_text', 'some_url', TRUE, FALSE),
    (3, 'some_title', 'some_text', 'some_url', TRUE, FALSE),
    (4, 'some_title', 'some_text', 'some_url', TRUE, FALSE),
    (5, 'some_title', 'some_text', 'some_url', TRUE, FALSE),
    (1, 'some_title', 'some_text', 'some_url', FALSE, TRUE),
    (2, 'some_title', 'some_text', 'some_url', FALSE, TRUE),
    (3, 'some_title', 'some_text', 'some_url', FALSE, TRUE),
    (4, 'some_title', 'some_text', 'some_url', FALSE, TRUE),
    (5, 'some_title', 'some_text', 'some_url', FALSE, TRUE)
ON CONFLICT
    DO NOTHING;

INSERT INTO users (name, ROLE)
    VALUES ('admin', 'ADMIN'),
    ('user', 'USER')
ON CONFLICT
    DO NOTHING;

INSERT INTO tags (tag_id, banner_id)
    VALUES (1, 1),
    (2, 1),
    (3, 5),
    (1, 2),
    (2, 3),
    (1, 6),
    (2, 7),
    (3, 8),
    (1, 9),
    (2, 10),
    (5, 6),
    (4, 7),
    (2, 9),
    (1, 7),
    (1, 8),
    (1, 10)
ON CONFLICT
    DO NOTHING;

