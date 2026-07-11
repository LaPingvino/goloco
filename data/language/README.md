# Language packs

These `*.yml` language packs are copied verbatim from the
[OpenLoco](https://github.com/OpenLoco/OpenLoco) project (`data/language/`) and
are used by `pkg/i18n` for goloco's UI localisation. They are OpenLoco's work,
distributed under the MIT License; all credit to the OpenLoco contributors.

Format: a `header:` block (`locale`, `native_name`, `loco_original_id`, …)
followed by a flat `strings:` table of `id: text` entries, where `text` may
contain `{FORMAT}` codes. goloco ships them so it can run standalone; the loader
also falls back to `~/OpenLoco/data/language`.
