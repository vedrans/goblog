-- ----------------------------
--  Sequence structure for article_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."article_id_seq";
CREATE SEQUENCE "public"."article_id_seq" INCREMENT 1 START 1 MAXVALUE 9223372036854775807 MINVALUE 1 CACHE 1;

-- ----------------------------
--  Table structure for article
-- ----------------------------
DROP TABLE IF EXISTS "public"."article";
CREATE TABLE "article" (
	"id" int4 NOT NULL DEFAULT nextval('article_id_seq'::regclass),
	"title" varchar(150) NOT NULL COLLATE "default",
	"content" text NOT NULL
)
WITH (OIDS=FALSE);

-- ----------------------------
--  Primary key structure for table article
-- ----------------------------
ALTER TABLE "public"."article" ADD PRIMARY KEY ("id") NOT DEFERRABLE INITIALLY IMMEDIATE;

-- ----------------------------
--  Sequence structure for appuser_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."appuser_id_seq";
CREATE SEQUENCE "public"."appuser_id_seq" INCREMENT 1 START 1 MAXVALUE 9223372036854775807 MINVALUE 1 CACHE 1;

-- ----------------------------
--  Table structure for appuser
-- ----------------------------
DROP TABLE IF EXISTS "public"."appuser";
CREATE TABLE "public"."appuser" (
	"id" int4 NOT NULL DEFAULT nextval('appuser_id_seq'::regclass),
	"name" varchar(255) NOT NULL COLLATE "default",
	"email" varchar(40) NOT NULL COLLATE "default",
	"password" character(60) NOT NULL
)
WITH (OIDS=FALSE);

-- ----------------------------
--  Primary key structure for table users
-- ----------------------------
ALTER TABLE "public"."appuser" ADD PRIMARY KEY ("id") NOT DEFERRABLE INITIALLY IMMEDIATE;
