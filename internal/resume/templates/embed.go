package templates

import "embed"

// FS 内置简历模板、皮肤与图标资产（禁止用户上传路径）。
//
//go:embed manifest.json themeTokens.json lucide-icons.json
//go:embed layouts/*.html base/*.css skins/*.css
var FS embed.FS
