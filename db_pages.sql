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
