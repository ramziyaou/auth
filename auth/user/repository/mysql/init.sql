CREATE TABLE IF NOT EXISTS `users`
(
    id bigint auto_increment,
    ts TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    iin varchar(255) NOT NULL UNIQUE,
    username varchar(255) NOT NULL UNIQUE,
    password varchar(255) NOT NULL,
    PRIMARY KEY (`id`)
);

INSERT INTO users (`iin`, `username`, `password`, `ts`) 
VALUES 
    ('0','admin', '$2a$10$YVWoFp84S4F7TkIkV2KhguNmQ4bkQRhN14fz.MeocFLOO7XBkLxH.', '2021-12-07 14:05:23'), -- password is 'password '
    ('910815450350', 'a', '$2a$10$fygvHR0NpECM.rKeIWtSYuL6SNY8SZEs83jWiUji5LPFYzLT6MAdO', '2021-12-07 14:01:03'),
    ('601119400567', 'm', '$2a$10$fygvHR0NpECM.rKeIWtSYuL6SNY8SZEs83jWiUji5LPFYzLT6MAdO', '2021-12-08 16:56:29'),
    ('980124450084', 'r', '$2a$10$fygvHR0NpECM.rKeIWtSYuL6SNY8SZEs83jWiUji5LPFYzLT6MAdO', '2021-12-25 02:59:41'),
    ('980124450072', 'mr', '$2a$10$fygvHR0NpECM.rKeIWtSYuL6SNY8SZEs83jWiUji5LPFYzLT6MAdO', '2022-01-13 19:45:20');