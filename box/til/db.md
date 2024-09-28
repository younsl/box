# 스터디 숙제

## Q1

거래소 서비스 오픈 전 테이블을 설계하는 단계입니다. 다음 두 테이블의 구조를 파악하고 문제가 있다면 수정을 해 주세요, 수정은 자유롭게 진행하시고, 수정 이후 각 컬럼과 수정에 대한 코멘트가 있으면 좋습니다.
FK, PK, 정규화, 인덱스 생성-추가, 파티션 등등 모든 의견은 자유롭게 반영 해 주세요.
SELECT UPDATE 시 주로 사용하는 조건절에는 email_id, mobile_number, last_login, is_mobile_auth 등 입니다.
테이블에 입력될 가상자산의 종류는 비트코인과 이더리움, 이더리움 클래식 외에 다른 코인이 추가될 수 있습니다.
만약 테이블 설계 변경 필요하다고 판단된다면, 잔고 데이터(*_balance)과 사용자 정보 간 서비스 영향도를 최소화하는 방식으로 고려해야 합니다.

```sql
create table coinone_user -- 고객정보와 고객 보유 코인 정보를 저장하는 테이블
 (
 email_id varchar(60) not null,
 passwd varchar(60) not null,
 user_name varchar(60) not null,
 mobile_number varchar(16) not null,
 home_address_1 varchar(128) default null, -- 집주소 읍면동
 home_address_2 varchar(128) default null, -- 집주소 번지, 호수
 birthday varchar(8) not null, -- 생일 yyyymmdd
 is_mobile_auth int default 0, -- 모바일 인증여부
 is_email_auth int default 0, -- 이메일 인증 여부
 is_account_auth int default 0, -- 계좌 인증 여부
 last_login varchar(8) not null, -- 마지막 로그인 시간
 login_ip varchar(16) not null, -- 최근 로그인 시 사용된 ip 주소
 btc_balance int null, -- 비트코인 잔고
 btc_wallet_address varchar(128) null, -- 비트코인 지갑 주소
 etc_balance int null, -- 이더리움 클래식 잔고
 etc_wallet_address varchar(128) null, -- 이더리움 클래식 지갑주소
 eth_balance int null, -- 이더리움 잔고
 eth_wallet_address varchar(128) null, -- 이더리움 지갑 주소
 primary key (email_id)
 ) engine = innodb default charset = utf8;
 
create table coinone_user_balance_history -- 고객의 자산 변경 이력을 저장하는 테이블. select 혹은 update 시 조건절에서 주로 사용하는 컬럼은 id, email_id, dt, coin_type 입니다.
(
 id int not null,
 email_id varchar(60) not null, -- 회원 아이디
 dt timestamp default now(), -- 시간
 coin_type varchar (8) not null, -- 자산 종류 코인 심볼 ex : etc, btc, eth
 balance int not null, -- 변경 당시 잔고
 primary key (id)
) engine = innodb default charset = utf8;
```

## A1

### 수정 사항 정리

- `user_id` 추가 및 PRIMARY KEY로 설정: 기존 `email_id`를 PK에서 제외하고 고유한 `user_id`를 `AUTO_INCREMENT`로 추가하여 PK로 설정.
- `email_id`에 UNIQUE 제약 조건 추가: 이메일 주소는 중복되지 않도록 UNIQUE 제약을 추가.
- 잔고 테이블 및 이력 테이블의 외래키 참조를 `user_id`로 변경: 기존 `email_id`를 참조하는 구조에서 `user_id`를 참조하도록 변경.
- 주요 조건에 대한 인덱스 추가: 주로 사용하는 SELECT/UPDATE 조건(mobile_number, last_login, is_mobile_auth)에 인덱스 추가.

### 1. coinone_user 테이블 (고유 user_id 추가 및 email_id의 UNIQUE 설정)

```sql
CREATE TABLE coinone_user -- 고객 정보를 저장하는 테이블
(
    user_id INT NOT NULL AUTO_INCREMENT, -- (1) 고유한 ID
    email_id VARCHAR(60) NOT NULL, -- 이메일 ID
    passwd VARCHAR(255) NOT NULL, -- 비밀번호는 해시된 값으로 저장
    user_name VARCHAR(60) NOT NULL, -- 사용자 이름
    mobile_number VARCHAR(16) NOT NULL, -- 모바일 번호
    home_address_1 VARCHAR(128) DEFAULT NULL, -- 집주소 1
    home_address_2 VARCHAR(128) DEFAULT NULL, -- 집주소 2
    birthday DATE NOT NULL, -- (2) 생년월일을 DATE 타입으로 변경
    is_mobile_auth TINYINT DEFAULT 0, -- 모바일 인증 여부
    is_email_auth TINYINT DEFAULT 0, -- 이메일 인증 여부
    is_account_auth TINYINT DEFAULT 0, -- 계좌 인증 여부
    last_login DATETIME NOT NULL, -- (3) 마지막 로그인 시간을 DATETIME으로 변경
    login_ip VARCHAR(45) NOT NULL, -- (4) IPv6를 고려하여 VARCHAR(45)로 변경
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- 생성 시간
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, -- 마지막 업데이트 시간
    PRIMARY KEY (user_id), -- (1) user_id를 PK로 설정
    UNIQUE (email_id), -- (2) email_id를 UNIQUE로 설정
    UNIQUE (mobile_number), -- 모바일 번호는 중복이 불가하므로 UNIQUE 제약 조건 추가
    INDEX idx_mobile_number (mobile_number), -- (5) SELECT/UPDATE에 자주 사용되므로 인덱스 추가
    INDEX idx_last_login (last_login), -- (5) SELECT/UPDATE에 자주 사용되므로 인덱스 추가
    INDEX idx_is_mobile_auth (is_mobile_auth) -- (5) SELECT/UPDATE에 자주 사용되므로 인덱스 추가
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 2. coinone_user_balance 테이블 (잔고 테이블에서 user_id 참조로 변경)

고객 잔고 정보를 저장하는 `coinone_user_balance` 테이블에서 `email_id` 대신 `user_id`를 참조하도록 수정했습니다. 이를 통해 고객의 이메일이 변경되더라도 잔고 데이터를 안전하게 유지할 수 있습니다.

```sql
CREATE TABLE coinone_user_balance -- 고객 잔고 정보를 저장하는 테이블
(
    user_id INT NOT NULL, -- (1) user_id를 참조
    coin_type VARCHAR(10) NOT NULL, -- 코인 심볼 (btc, eth, etc 등)
    balance DECIMAL(20, 8) NOT NULL, -- 잔고는 소수점 처리가 필요하므로 DECIMAL로 변경
    wallet_address VARCHAR(128) NOT NULL, -- 코인 지갑 주소
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- 생성 시간
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, -- 업데이트 시간
    PRIMARY KEY (user_id, coin_type), -- user_id와 코인 심볼로 복합키 구성
    FOREIGN KEY (user_id) REFERENCES coinone_user(user_id) ON DELETE CASCADE, -- (2) user_id로 FK 설정
    INDEX idx_user_coin (user_id, coin_type) -- (3) 자주 사용하는 조건에 대한 인덱스 추가
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 3. coinone_user_balance_history 테이블 (user_id 참조로 변경)

이력 테이블에서도 `user_id`를 참조하도록 수정하여, 고객이 이메일을 변경하더라도 자산 변경 이력이 올바르게 관리되도록 설정합니다.

```sql
CREATE TABLE coinone_user_balance_history -- 고객의 자산 변경 이력을 저장하는 테이블
(
    id INT NOT NULL AUTO_INCREMENT, -- ID 자동 증가
    user_id INT NOT NULL, -- (1) user_id를 참조
    dt TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- 변경 시각
    coin_type VARCHAR(10) NOT NULL, -- 코인 심볼
    balance DECIMAL(20, 8) NOT NULL, -- 잔고는 소수점 처리
    PRIMARY KEY (id), -- 고유 ID
    FOREIGN KEY (user_id) REFERENCES coinone_user(user_id) ON DELETE CASCADE, -- (2) user_id로 FK 설정
    INDEX idx_user_dt (user_id, dt), -- (3) 자주 사용하는 조건에 대한 인덱스 추가
    INDEX idx_coin_type (coin_type) -- (3) 코인 종류에 대한 인덱스 추가
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 최종 정리:

- `user_id`를 각 테이블의 PK로 사용하고, 각 테이블이 이를 참조하도록 수정하였습니다.
- `email_id`는 UNIQUE로 설정하여 고객의 이메일이 변경되더라도 유저 식별이 가능한 구조로 변경하였습니다.
- 인덱스 추가: 주로 사용하는 조건(mobile_number, last_login, is_mobile_auth, coin_type)에 대한 인덱스를 추가하여 성능을 최적화하였습니다.

이 설계를 통해 고객 데이터 관리가 더 안정적이고 유연해졌으며, 확장성도 확보 가능하게 개선됨.

## Q2

`coinone_user_balance_history` 테이블에서 특정일자의 가장 마지막 자산 변경 이력을 조회하는 SQL문을 작성해 주세요.
만약 `coinone_user_balance_history` 테이블을 수정하셨다면 수정된 버전 기준으로 작성해 주시면 됩니다.
필요한 인덱스는 추가로 설계 가능합니다.
특정일자의 한정은 `dt` 기준으로 `BETWEEN` 절을 사용하여 24시간으로 한정 하시고 24시간 동안 발생한 자산 변경이력의 가장 마지막 이력이 각 코인별로, 회원별로 출력되는것이 목표 입니다.

## A2

coinone_user_balance_history 테이블에서 특정일자의 가장 마지막 자산 변경 이력을 조회하는 SQL을 작성합니다.

### SQL 작성 목표

- 특정일자의 24시간 내에서 발생한 자산 변경 이력 중, 각 코인별, 회원별로 가장 마지막 이력을 조회.
- dt 컬럼을 기준으로 BETWEEN을 사용하여 24시간 내로 한정.
- user_id와 coin_type별로 가장 최근의 자산 변경 이력을 가져오도록 쿼리 작성.

### 필요한 인덱스 추가

자주 사용하는 컬럼인 `user_id`, `coin_type`, `dt`에 대해 성능을 최적화하기 위해 인덱스를 추가.

```sql
-- 인덱스 추가: user_id, coin_type, dt를 기준으로 조회 성능 향상
CREATE INDEX idx_user_coin_dt ON coinone_user_balance_history (user_id, coin_type, dt);
```

아래 SQL은 BETWEEN 절을 사용하여 특정일의 24시간 동안 발생한 각 회원의 각 코인별 가장 마지막 자산 변경 이력을 조회하는 쿼리입니다.

```sql
SELECT 
    h.user_id, 
    h.coin_type, 
    h.balance, 
    h.dt
FROM 
    coinone_user_balance_history h
INNER JOIN (
    -- (1) 각 user_id, coin_type 별 가장 최근 dt를 찾는 서브쿼리
    SELECT 
        user_id, 
        coin_type, 
        MAX(dt) AS last_dt
    FROM 
        coinone_user_balance_history
    WHERE 
        dt BETWEEN '2024-09-25 00:00:00' AND '2024-09-25 23:59:59' -- (2) 24시간 동안의 데이터 한정
    GROUP BY 
        user_id, coin_type -- (3) 회원별, 코인별 그룹화
) recent ON 
    h.user_id = recent.user_id 
    AND h.coin_type = recent.coin_type 
    AND h.dt = recent.last_dt -- (4) 가장 최근 dt와 일치하는 레코드 선택
ORDER BY 
    h.user_id, h.coin_type;
```

### 쿼리 설명

- 서브쿼리에서 `user_id`, `coin_type`별로 가장 최근의 dt를 찾습니다. 이때 BETWEEN 절을 사용해 특정일의 24시간 범위를 지정합니다.
- `dt BETWEEN '2024-09-25 00:00:00' AND '2024-09-25 23:59:59'`로 24시간의 기간을 설정하여, 해당 기간에 발생한 이력만 조회합니다.
- `GROUP BY user_id, coin_type`으로 각 회원의 코인별로 그룹화하여, 각 코인별 가장 마지막 이력을 가져옵니다.
- 메인 쿼리에서 이 서브쿼리 결과와 coinone_user_balance_history 테이블을 조인하여, 각 회원과 코인별로 마지막 이력(MAX(dt))을 기준으로 레코드를 가져옵니다.

| user_id | coin_type | balance | dt                  |
|---------|-----------|---------|---------------------|
| 101     | btc       | 0.6     | 2024-09-25 12:00:00 |  <!-- user_id 101의 btc에서 가장 최근 이력 -->
| 101     | eth       | 1.2     | 2024-09-25 16:00:00 |  <!-- user_id 101의 eth에서 가장 최근 이력 -->
| 102     | btc       | 0.3     | 2024-09-25 17:30:00 |  <!-- user_id 102의 btc에서 가장 최근 이력 -->
| 102     | eth       | 0.5     | 2024-09-25 15:00:00 |  <!-- user_id 102의 eth에서 가장 최근 이력 -->

### 성능 최적화

쿼리의 성능을 최적화하기 위해 user_id, coin_type, dt에 대해 복합 인덱스를 추가하였습니다. 이 인덱스를 사용하면 WHERE, GROUP BY와 JOIN에서 사용되는 컬럼들이 효과적으로 인덱스를 타기 때문에 성능이 개선됩니다.
