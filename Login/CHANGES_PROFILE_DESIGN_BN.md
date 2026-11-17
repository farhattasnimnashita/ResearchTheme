# Profile UI update

- Student Profile: profile accent colours now match the Student Dashboard blue palette instead of the previous green palette.
- Teacher Profile: redesigned the profile overview to use the same information hierarchy as the student profile (identity heading, photo/overview hero, information cards, edit action), while preserving the teacher dashboard's blue theme and existing edit/save/photo-upload handlers.
- Both profile views have responsive layouts for smaller screens.
- No database schema or profile API changes were made.

## Validation
- `node --check` passed for the updated teacher profile view and dashboard JavaScript.
- Full Go tests could not run because Go dependencies were not available and network access to `proxy.golang.org` timed out.
