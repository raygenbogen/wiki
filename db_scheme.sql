-- Table: pages

-- DROP TABLE pages;

CREATE TABLE pages
(
  title character varying NOT NULL,
  author character varying,
  "timestamp" character varying NOT NULL,
  content text,
  CONSTRAINT pages_pkey PRIMARY KEY (title, "timestamp")
)
WITH (
  OIDS=FALSE
);
ALTER TABLE pages
  OWNER TO postgres;

-- Table: users

-- DROP TABLE users;

CREATE TABLE users
(
	  name character varying NOT NULL,
	  approved character varying,
	  admin character varying,
	  password character varying,
	  CONSTRAINT users_pkey PRIMARY KEY (name)
)
WITH (
	  OIDS=FALSE
);
ALTER TABLE users
  OWNER TO postgres;
