CREATE TABLE star (
    id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY,
    keyword VARCHAR(191) NOT NULL,
    user_name VARCHAR(191) NOT NULL,
    created_at DATETIME,
    INDEX keyword_idx(keyword)
) Engine=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
