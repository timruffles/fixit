-- Create "communities" table
CREATE TABLE "communities" (
  "id" bigserial NOT NULL,
  "name" character varying NOT NULL,
  "title" character varying NOT NULL,
  PRIMARY KEY ("id")
);