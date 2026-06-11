[Español](README.es.md) | [English](README.en.md)

---

# Core Utils CLI (`cu`)

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go](https://img.shields.io/badge/go-1.22+-00ADD8.svg)
![Status](https://img.shields.io/badge/status-stable-green.svg)

**Core Utils CLI (`cu`)** es el gestor de paquetes / lanzador del ecosistema [Core Utils](https://core-utils.dev). Se conecta al registry de Core Utils para buscar, instalar, actualizar y eliminar las herramientas CLI, scripts y apps publicados en [core-utils.dev](https://core-utils.dev).

---

## ⚠️ La instalación es distinta al resto de herramientas de Core Utils

La mayoría de proyectos de Core Utils se instalan desde el código fuente (`git clone` + `pipx install` / `init.sh`). **`core-utils-cli` es la excepción.**

Como `cu` es la herramienta que se usa *para instalar todo lo demás*, se distribuye como un **binario precompilado** mediante **GitHub Releases**. No hace falta `pip install`, ni `pipx`, ni tener Go instalado para usarlo — solo descarga el build correspondiente a tu sistema operativo y ejecútalo.

### 1. Descarga el archivo correcto para tu sistema

Entra a la [última release](https://github.com/mdwcoder/core-utils-cli/releases/latest) y descarga el archivo según tu sistema operativo:

| Sistema operativo | Archivo | Notas |
|---|---|---|
| Linux (x86_64) | `cu-linux-amd64.AppImage` | Autocontenido, se ejecuta directamente. Puede requerir `--appimage-extract-and-run` o FUSE. |
| Linux (x86_64) | `cu-linux-amd64.zip` | Zip simple con el binario `cu`. Recomendado para servidores / scripting. |
| macOS (Intel + Apple Silicon) | `cu-macos-universal.dmg` | Binario universal (amd64 + arm64). Sin firmar — ver notas abajo. |
| Windows (x86_64) | `cu-windows-amd64.exe` | Sin firmar — Windows SmartScreen puede avisar la primera vez. |

Cada release publica también un archivo `checksums.txt` con el SHA256 de cada asset.

### 2. Instálalo

**Linux (recomendado, desde el zip):**

```bash
curl -fsSL -o cu.zip https://github.com/mdwcoder/core-utils-cli/releases/latest/download/cu-linux-amd64.zip
unzip cu.zip -d /tmp/cu-extract
install -Dm755 /tmp/cu-extract/cu ~/.local/bin/cu
```

Asegúrate de que `~/.local/bin` esté en tu `PATH`, y verifica con:

```bash
cu version
cu doctor
```

También puedes usar el script de ayuda, que hace los pasos anteriores automáticamente (Linux/macOS):

```bash
curl -fsSL https://raw.githubusercontent.com/mdwcoder/core-utils-cli/main/scripts/install.sh | bash
```

**Linux (AppImage):**

```bash
curl -fsSL -o cu.AppImage https://github.com/mdwcoder/core-utils-cli/releases/latest/download/cu-linux-amd64.AppImage
chmod +x cu.AppImage
./cu.AppImage version
```

**macOS:**

1. Descarga `cu-macos-universal.dmg` desde la [última release](https://github.com/mdwcoder/core-utils-cli/releases/latest).
2. Abre el `.dmg` y copia `cu` a `/usr/local/bin` (o cualquier carpeta dentro de tu `PATH`).
3. Como el binario no está firmado, la primera vez que lo ejecutes Gatekeeper puede bloquearlo. Haz clic derecho sobre el binario → **Abrir**, o ejecuta:
   ```bash
   xattr -d com.apple.quarantine /usr/local/bin/cu
   ```

**Windows:**

1. Descarga `cu-windows-amd64.exe` desde la [última release](https://github.com/mdwcoder/core-utils-cli/releases/latest).
2. Muévelo a una carpeta de tu elección y añade esa carpeta a tu `PATH`, o ejecútalo directamente como `cu.exe`.
3. El ejecutable no está firmado — Windows SmartScreen puede mostrar "Windows protegió tu PC". Haz clic en **Más información → Ejecutar de todas formas**.

### 3. Verifica el checksum (recomendado)

```bash
curl -fsSLO https://github.com/mdwcoder/core-utils-cli/releases/latest/download/checksums.txt
sha256sum -c checksums.txt --ignore-missing
```

---

## Uso

```text
Core Utils CLI (cu) — Uso:

  cu help                  Muestra esta ayuda
  cu version               Muestra la versión del CLI
  cu registry sync         Descarga y guarda en caché el registry
  cu search <query>        Busca herramientas en el registry
  cu info <tool>           Muestra detalles de una herramienta
  cu install <tool>        Instala una herramienta
  cu list                  Lista las herramientas instaladas
  cu update [tool]         Actualiza todas o una herramienta concreta
  cu remove <tool>         Elimina una herramienta
  cu doctor                Diagnostica el entorno

Variables de entorno:
  CORE_UTILS_REGISTRY_URL  Sobrescribe el endpoint del registry
  NO_COLOR                 Desactiva la salida con color
```

### Ejemplos

```bash
# Sincronizar el registry
cu registry sync

# Buscar herramientas
cu search note

# Ver detalles de una herramienta
cu info memory-note-cli

# Instalar una herramienta para tu plataforma
cu install memory-note-cli

# Listar herramientas instaladas
cu list

# Actualizar todo (o solo una herramienta)
cu update
cu update memory-note-cli

# Eliminar una herramienta
cu remove memory-note-cli

# Comprobar tu entorno
cu doctor
```

`cu` guarda sus datos en `~/.core-utils/` (`bin/`, `tools/<tool>/<version>/`, `config.json`, `installed.json`, `registry-cache.json`).

---

## Cómo funciona `cu install`

1. Resuelve la herramienta por id o nombre (sin distinguir mayúsculas/minúsculas) en el registry remoto o en la caché local.
2. Detecta tu plataforma (`GOOS`/`GOARCH` → `linux-amd64`, `darwin-arm64`, etc.).
3. Descarga el asset correspondiente a un archivo temporal.
4. Valida su checksum SHA256 cuando está disponible.
5. Extrae el archive (zip / tar.gz) con protección contra path traversal.
6. Instala el binario en `~/.core-utils/bin/`.
7. Registra la instalación en `~/.core-utils/installed.json`.

Si `~/.core-utils/bin` no está en tu `PATH`, `cu doctor` y `cu install` te avisarán.

---

## Releases y versionado

Las releases se generan y publican automáticamente mediante GitHub Actions cada vez que se sube un tag con el formato `v*.*.*`. Cada release incluye:

- AppImage y zip para Linux (amd64)
- DMG universal para macOS (amd64 + arm64)
- EXE para Windows (amd64)
- `checksums.txt` con los SHA256 de todos los assets

`cu version` muestra la versión, commit y fecha de build inyectados en tiempo de compilación:

```text
Core Utils CLI
Version:   v0.2.0
Commit:    8176ac2...
Build date: 2026-06-11T04:13:51Z
Platform:  linux/amd64
```

> **Nota:** los builds de macOS y Windows actualmente **no están firmados** ni notarizados. Es algo esperado en un proyecto open-source en etapa temprana — verifica siempre el checksum si tienes dudas.

Consulta [`RELEASE_PLAN.md`](RELEASE_PLAN.md) para el proceso completo de releases.

---

## Compilar desde el código fuente

Si prefieres compilarlo tú mismo:

```bash
git clone https://github.com/mdwcoder/core-utils-cli.git
cd core-utils-cli
./scripts/build.sh
./bin/cu version
```

Requiere Go 1.22+.

---

## Licencia

MIT
