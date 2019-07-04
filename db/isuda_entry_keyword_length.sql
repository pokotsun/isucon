ALTER TABLE entry ADD keyword_length BIGINT NOT NULL;
UPDATE entry SET keyword_length=CHARACTER_LENGTH(keyword);
