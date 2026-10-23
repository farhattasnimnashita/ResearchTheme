-- Run once in your shared PostgreSQL/Supabase database.
-- Directory entries only: this creates no login accounts or approvals.
BEGIN;
CREATE SCHEMA IF NOT EXISTS researchbridge;
CREATE TABLE IF NOT EXISTS researchbridge.teacher_directory (
 id TEXT PRIMARY KEY,
 name TEXT NOT NULL,
 designation TEXT NOT NULL,
 institution TEXT NOT NULL,
 department TEXT NOT NULL,
 phone TEXT,
 email TEXT NOT NULL UNIQUE,
 website_url TEXT,
 photo_url TEXT,
 research_interests JSONB NOT NULL DEFAULT '[]'::jsonb,
 expertise JSONB NOT NULL DEFAULT '[]'::jsonb,
 accepting_students BOOLEAN,
 source TEXT NOT NULL,
 verified BOOLEAN NOT NULL DEFAULT FALS
);
-- Keep database access behind the existing authenticated Go API.
ALTER TABLE researchbridge.teacher_directory ENABLE ROW LEVEL SECURITY;
INSERT INTO researchbridge.teacher_directory (id, name, designation, institution, department, phone, email, source) VALUES ('sust-cse-001', 'Mohammad Shahidur Rahman, PhD', 'Professor', 'Shahjalal University of Science and Technology', 'Computer Science & Engineering', '+8801914930807', 'rahmanms@sust.edu', 'User-provided directory screenshots dated 2026-10-18') ON CONFLICT DO NOTHING;
INSERT INTO researchbridge.teacher_directory (id, name, designation, institution, department, phone, email, source) VALUES ('sust-cse-002', 'Dr Mohammad Reza Selim', 'Professor', 'Shahjalal University of Science and Technology', 'Computer Science & Engineering', '+8801972357830', 'selim@sust.edu', 'User-provided directory screenshots dated 2026-10-18') ON CONFLICT DO NOTHING;
INSERT INTO researchbridge.teacher_directory (id, name, designation, institution, department, phone, email, source) VALUES ('sust-cse-003', 'M. Jahirul Islam, PhD., PEng.', 'Professor', 'Shahjalal University of Science and Technology', 'Computer Science & Engineering', '+8801770348185', 'jahir-cse@sust.edu', 'User-provided directory screenshots dated 2026-10-18') ON CONFLICT DO NOTHING;
INSERT INTO researchbridge.teacher_directory (id, name, designation, institution, department, phone, email, source) VALUES ('sust-cse-004', 'Mohammad Abdullah Al Mumin, PhD', 'Professor', 'Shahjalal University of Science and Technology', 'Computer Science & Engineering', '+8801711445110', 'mumin-cse@sust.edu', 'User-provided directory screenshots dated 2026-10-18') ON CONFLICT DO NOTHING;
INSERT INTO researchbridge.teacher_directory (id, name, designation, institution, department, phone, email, source) VALUES ('sust-cse-005', 'Dr. Md Forhad Rabbi, SMIEEE', 'Professor & Head', 'Shahjalal University of Science and Technology', 'Computer Science & Engineering', '+8801844175805', 'frabbi-cse@sust.edu', 'User-provided directory screenshots dated 2026-10-18') ON CONFLICT DO NOTHING;
INSERT INTO researchbridge.teacher_directory (id, name, designation, institution, department, phone, email, source) VALUES ('sust-cse-006', 'Md Masum', 'Professor', 'Shahjalal University of Science and Technology', 'Computer Science & Engineering', '+8801919736248', 'masum-cse@sust.edu', 'User-provided directory screenshots dated 2026-10-18') ON CONFLICT DO NOTHING;
INSERT INTO researchbridge.teacher_directory (id, name, designation, institution, department, phone, email, source) VALUES ('sust-cse-007', 'Dr. Husne Ara Chowdhury', 'Professor', 'Shahjalal University of Science and Technology', 'Computer Science & Engineering', '+8801558304163', 'husna-cse@sust.edu', 'User-provided directory screenshots dated 2026-10-18') ON CONFLICT DO NOTHING;
INSERT INTO researchbridge.teacher_directory (id, name, designation, institution, department, phone, email, source) VALUES ('sust-cse-008', 'Dr. Sadia Sultana', 'Professor', 'Shahjalal University of Science and Technology', 'Computer Science & Engineering', '+8801911089612', 'sadia-cse@sust.edu', 'User-provided directory screenshots dated 2026-10-18') ON CONFLICT DO NOTHING;
INSERT INTO researchbridge.teacher_directory (id, name, designation, institution, department, phone, email, source) VALUES ('sust-cse-009', 'Mahruba Sharmin Chowdhury', 'Associate Professor', 'Shahjalal University of Science and Technology', 'Computer Science & Engineering', '+8801917566699', 'mahruba-cse@sust.edu', 'User-provided directory screenshots dated 2026-10-18') ON CONFLICT DO NOTHING;
INSERT INTO researchbridge.teacher_directory (id, name, designation, institution, department, phone, email, source) VALUES ('sust-cse-010', 'Md. Eamin Rahman', 'Assistant Professor', 'Shahjalal University of Science and Technology', 'Computer Science & Engineering', '+8801677014633', 'eamin-cse@sust.edu', 'User-provided directory screenshots dated 2026-10-18') ON CONFLICT DO NOTHING;
INSERT INTO researchbridge.teacher_directory (id, name, designation, institution, department, phone, email, source) VALUES ('sust-cse-011', 'A.K.M. Fakhrul Hossain', 'Lecturer', 'Shahjalal University of Science and Technology', 'Computer Science & Engineering', '+8801521333411', 'fakhrul-cse@sust.edu', 'User-provided directory screenshots dated 2026-10-18') ON CONFLICT DO NOTHING;
INSERT INTO researchbridge.teacher_directory (id, name, designation, institution, department, phone, email, source) VALUES ('sust-cse-012', 'Md. Shadmim Hasan Sifat', 'Lecturer', 'Shahjalal University of Science and Technology', 'Computer Science & Engineering', '+8801571291835', 'shadmim-cse@sust.edu', 'User-provided directory screenshots dated 2026-10-18') ON CONFLICT DO NOTHING;
INSERT INTO researchbridge.teacher_directory (id, name, designation, institution, department, phone, email, source) VALUES ('sust-cse-013', 'Md. Mehedi Hasan', 'Lecturer', 'Shahjalal University of Science and Technology', 'Computer Science & Engineering', '+8801932981532', 'mehedi-cse@sust.edu', 'User-provided directory screenshots dated 2026-10-18') ON CONFLICT DO NOTHING;
INSERT INTO researchbridge.teacher_directory (id, name, designation, institution, department, phone, email, source) VALUES ('sust-cse-014', 'Abdullah Al Thaki', 'Lecturer', 'Shahjalal University of Science and Technology', 'Computer Science & Engineering', '+8801516178219', 'toki-cse@sust.edu', 'User-provided directory screenshots dated 2026-10-18') ON CONFLICT DO NOTHING;
INSERT INTO researchbridge.teacher_directory (id, name, designation, institution, department, phone, email, source) VALUES ('sust-cse-015', 'Md. Shymon Islam', 'Lecturer', 'Shahjalal University of Science and Technology', 'Computer Science & Engineering', '+8801718550548', 'shymon-cse@sust.edu', 'User-provided directory screenshots dated 2026-10-18') ON CONFLICT DO NOTHING;
INSERT INTO researchbridge.teacher_directory (id, name, designation, institution, department, phone, email, source) VALUES ('sust-cse-016', 'MD Fahimul Islam Fahim', 'Lecturer', 'Shahjalal University of Science and Technology', 'Computer Science & Engineering', '+8801893639181', 'fahim-cse@sust.edu', 'User-provided directory screenshots dated 2026-10-18') ON CONFLICT DO NOTHING;
INSERT INTO researchbridge.teacher_directory (id, name, designation, institution, department, phone, email, source) VALUES ('sust-cse-017', 'Mainul Hasan Mohsin', 'Lecturer', 'Shahjalal University of Science and Technology', 'Computer Science & Engineering', '+8801521565271', 'mainul-cse@sust.edu', 'User-provided directory screenshots dated 2026-10-18') ON CONFLICT DO NOTHING;
INSERT INTO researchbridge.teacher_directory (id, name, designation, institution, department, phone, email, source) VALUES ('sust-cse-018', 'Mufassir Ahmad Chowdhury', 'Lecturer', 'Shahjalal University of Science and Technology', 'Computer Science & Engineering', '+8801771144308', 'mufassir-cse@sust.edu', 'User-provided directory screenshots dated 2026-10-18') ON CONFLICT DO NOTHING;
COMMIT;
