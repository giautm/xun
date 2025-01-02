# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
- use `html/template` to parse template files (#7)

## [1.0.3] - 2025-01-01
### Changed
- renamed package name with `xun` (#4)
- moved `htmx` helper to `ext/htmx` (#4)

### Fixed
- fixed syntax issue on `htmx.WriteHeader`
  
### Added
- added logging `app.routes` in `app.Start`

## [1.0.1] - 2024-12-30
### Added
- added htmx helper
- support setup custom FuncMap on HtmlTemplate 

## [1.0.0] - 2024-12-25
