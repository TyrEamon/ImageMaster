# Source Manifests

`sources/` is the local external-source directory for ImageMaster.

Current layers:
- root `*.json`: manifests that ImageMaster can already wire into current built-in adapters/runtime
- `miru/`: imported Miru manga catalog manifests for future adaptation
- `scripts/`: local executable JS source files loaded by ImageMaster source runtime

What these manifests do now:
- describe source name, language, capabilities, website, and origin metadata
- optionally map a manifest to an existing ImageMaster adapter
- optionally point to a local `script` file and run it through the ImageMaster JS source runtime
- prepare the directory layout for a future remote source repository

What they do not do yet:
- execute Miru scripts directly without adaptation
- provide full Miru compatibility
- auto-install or auto-update themselves from a remote repo

Recommended workflow:
1. Keep working adapters in the root `sources/` directory.
2. Keep imported catalogs in `sources/miru/`.
3. When a Miru source is adapted for ImageMaster, either:
   - promote it into the root source index, or
   - wire the manifest to a new adapter/runtime and explicitly enable it.
4. Prefer new external sources to be added as `manifest + script` so future source updates do not require new Go code.
