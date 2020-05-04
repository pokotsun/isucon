ALTER TABLE entry ADD keyword_length BIGINT NOT NULL;
UPDATE entry SET keyword_length=CHARACTER_LENGTH(keyword);
ALTER TABLE entry ADD INDEX keyword_length_idx(keyword_length);
