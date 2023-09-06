
DROP PROCEDURE IF EXISTS gen_test_data;
DELIMITER $$
CREATE PROCEDURE gen_test_data (IN count INT UNSIGNED)
BEGIN
    IF count = '' THEN SET count = 1000000; END IF;

    SET @x = 0;
    REPEAT SET @x = @x + 1; 
        select rand()*1000000 into @num;
        insert into `test` values (null, @num, '');
    UNTIL @x >= count END REPEAT;

    SELECT count(*) FROM `test`;
END$$
DELIMITER ;


call gen_test_data(10000);
