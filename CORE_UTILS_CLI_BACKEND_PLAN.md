# Core Utils CLI + Backend Integration Plan

## 1. Resumen de la arquitectura actual de core-utils-backend

- **Framework**: FastAPI + Uvicorn
- **Base de datos**: SQLite (modo WAL) via `aiosqlite` + SQLAlchemy 2.0 async ORM
- **Estructura modular**:
  - `app/main.py` — punto de entrada, registra routers
  - `app/routers/` — rutas API (auth, threads, comments, votes, roadmap, tools, me, connect, relay, dev_profiles, easyforms, updates, admin)
  - `app/models/` — modelos SQLAlchemy
  - `app/schemas/` — Pydantic schemas
  - `app/services/` — lógica de negocio (media storage, optimizers, serializers)
  - `app/auth/` — JWT, GitHub OAuth, dependencias
  - `app/config.py` — Settings via pydantic-settings + `.env`
  - `app/database.py` — engine, session, init con WAL + migraciones lightweight
- **Deploy**: systemd en producción (`uvicorn app.main:app --host 0.0.0.0 --port 8000`), con Nginx/Caddy delante. Nada de Docker, Kubernetes, ni PostgreSQL.
- **Autenticación**: GitHub OAuth + JWT en cookie `forge_token` + header `Authorization: Bearer`
- **Rate limiting**: slowapi
- **CORS**: configurado por `CORS_ORIGINS`
- **Static data**: el router `/api/forge/tools` actualmente devuelve una lista estática de herramientas (sin releases ni assets).

## 2. Resumen de la arquitectura actual de core-utils-cli

- **Estado actual**: proyecto vacío excepto por `.gitignore` de Go.
- **Ubicación**: `core-utils-projects/core-utils-cli/`
- No hay código existente ni dependencias. Se debe crear desde cero en Go.

## 3. Cómo se va a organizar el CLI en Go

```
core-utils-cli/
├── go.mod
├── go.sum
├── README.md
├── CORE_UTILS_CLI_BACKEND_PLAN.md
├── cmd/
│   └── cu/
│       └── main.go
├── internal/
│   ├── cli/
│   │   └── commands.go      # registro de comandos y subcomandos
│   ├── registry/
│   │   ├── registry.go      # tipos JSON, fetch, cache
│   │   └── client.go        # HTTP client con timeout
│   ├── installer/
│   │   └── installer.go     # install, update, remove
│   ├── config/
│   │   └── config.go        # paths, config.json, env vars
│   ├── platform/
│   │   └── platform.go      # GOOS/GOARCH mapping
│   ├── checksum/
│   │   └── checksum.go      # sha256 validation
│   ├── archive/
│   │   └── archive.go       # zip / tar.gz extraction con path traversal protection
│   └── output/
│       └── output.go        # pretty printing, status OK/WARN/ERROR
├── scripts/
│   ├── build.sh
│   └── install-local.sh
└── tests/
    └── (unitarios en Go, _test.go dentro de cada package)
```

## 4. Qué endpoints necesita el backend

Todos bajo el prefijo `/api/registry` (nuevo router, sin tocar los existentes):

- `GET /api/registry` — registry completo con metadatos (`version`, `updatedAt`, `tools`)
- `GET /api/registry/tools` — alias / lista de tools
- `GET /api/registry/tools/{id}` — detalle de una tool
- `GET /api/registry/tools/{id}/releases/latest` — release más reciente con assets por plataforma

El backend ya usa `/api/forge/tools` para listar herramientas estáticas. El nuevo registry no elimina ni reemplaza ese endpoint; es complementario y orientado a la CLI (con versiones, assets, plataformas).

## 5. Cómo se conectará la CLI con el backend

- La CLI usa `net/http` con `context.WithTimeout` para todas las peticiones.
- URL base configurable por:
  1. Variable de entorno `CORE_UTILS_REGISTRY_URL` (prioridad máxima)
  2. Campo `registry_url` en `~/.core-utils/config.json`
  3. Valor por defecto: `https://core-utils.dev/api/registry`
- `cu registry sync` descarga `GET /api/registry` y guarda `registry-cache.json`.
- `cu search`, `cu info`, `cu install`, `cu update` consultan remoto primero; si falla, usan caché y muestran `WARN`.

## 6. Formato del registry

```json
{
  "version": 1,
  "updatedAt": "2026-06-11T00:00:00Z",
  "tools": [
    {
      "id": "memory-note-cli",
      "name": "MemoryNoteCLI",
      "description": "Fast terminal note manager",
      "category": "Productivity",
      "language": "Python",
      "version": "1.0.0",
      "binary": "memory-note",
      "homepage": "https://core-utils.dev",
      "repository": "https://github.com/mdwcoder/MemoryNoteCLI",
      "platforms": {
        "linux-amd64": {
          "type": "archive",
          "url": "TODO",
          "sha256": "CHANGE_ME",
          "binaryPath": "memory-note"
        },
        "windows-amd64": {
          "type": "archive",
          "url": "TODO",
          "sha256": "CHANGE_ME",
          "binaryPath": "memory-note.exe"
        }
      }
    }
  ]
}
```

- `sha256: "CHANGE_ME"` indica explícitamente que no hay checksum real todavía.
- `url: "TODO"` indica que no hay asset publicado aún.
- El CLI mostrará `WARN` cuando vea `CHANGE_ME` o `TODO`.

## 7. Modelo de datos recomendado para tools/releases/assets

**Backend (SQLite, async SQLAlchemy)**:

```python
class RegistryTool(Base):
    __tablename__ = "registry_tools"
    id = Column(String(64), primary_key=True, index=True)
    name = Column(String(120), nullable=False)
    description = Column(Text, nullable=True)
    category = Column(String(40), nullable=True)
    language = Column(String(40), nullable=True)
    homepage = Column(String(500), nullable=True)
    repository = Column(String(500), nullable=True)
    binary = Column(String(120), nullable=True)
    created_at = Column(DateTime, server_default=func.now(), nullable=False)
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now(), nullable=False)

class RegistryRelease(Base):
    __tablename__ = "registry_releases"
    id = Column(Integer, primary_key=True, autoincrement=True)
    tool_id = Column(String(64), ForeignKey("registry_tools.id"), nullable=False, index=True)
    version = Column(String(40), nullable=False)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, server_default=func.now(), nullable=False)
    __table_args__ = (UniqueConstraint("tool_id", "version", name="uq_release_version"),)

class RegistryAsset(Base):
    __tablename__ = "registry_assets"
    id = Column(Integer, primary_key=True, autoincrement=True)
    release_id = Column(Integer, ForeignKey("registry_releases.id"), nullable=False, index=True)
    platform = Column(String(30), nullable=False)  # linux-amd64, darwin-arm64, etc.
    asset_type = Column(String(20), nullable=False, default="archive")  # archive | binary
    url = Column(String(800), nullable=False)
    sha256 = Column(String(64), nullable=True)
    binary_path = Column(String(200), nullable=True)
    created_at = Column(DateTime, server_default=func.now(), nullable=False)
```

**Alternativa inmediata**: como aún no hay releases reales, el backend puede funcionar con un JSON estático local (`data/registry.json`) servido por un router FastAPI. Esto evita alterar la base de datos hasta que sea necesario, y mantiene el backend 100% estable. Cuando haya releases reales, se migra a las tablas SQLAlchemy con un script lightweight.

**Decisión tomada**: implementar el registry como JSON estático local (`data/registry.json`) más un router FastAPI que lo lee y sirve. Es la base más sólida y menos intrusiva. Las tablas SQLAlchemy se dejan definidas en el modelo para extensión futura, pero no se usan todavía.

## 8. Dónde guardar datos en el backend según la estructura actual

- `data/registry.json` — archivo JSON estático del registry.
- `data/` ya existe y contiene `core_utils_forge.db` y `uploads/`. Es el lugar natural.
- No se crea ninguna nueva tabla activa ahora; si en el futuro se migra a DB, el `init_db` ya creará las tablas automáticamente porque `Base.metadata.create_all` las descubrirá.

## 9. Cómo mantener compatibilidad con el sistema de deploy existente

- No se toca `.env.example`, `.env`, systemd service, ni scripts de deploy.
- Solo se añade:
  - `app/routers/registry.py` (nuevo router)
  - `app/schemas/registry.py` (nuevo schema)
  - `app/models/registry.py` (modelos definidos pero no activos todavía)
  - `data/registry.json` (datos estáticos)
  - Línea en `app/main.py` para incluir el router: `app.include_router(registry.router, prefix="/api")`
- No se añaden variables de entorno obligatorias. Si se necesita alguna opcional, se documenta con fallback.

## 10. Comandos del CLI

| Comando | Descripción |
|---------|-------------|
| `cu help` | Ayuda general |
| `cu version` | Versión del CLI |
| `cu registry sync` | Descarga y cachea el registry |
| `cu search <query>` | Busca en id, name, description, category, language |
| `cu info <tool>` | Muestra detalles de una tool |
| `cu install <tool>` | Descarga, valida, extrae, instala en `~/.core-utils/` |
| `cu list` | Lista herramientas instaladas |
| `cu update [tool]` | Actualiza todo o una tool específica |
| `cu remove <tool>` | Elimina una tool instalada |
| `cu doctor` | Diagnóstico completo del entorno |

## 11. Medidas de seguridad

- Solo HTTPS en producción; HTTP permitido solo en local/dev.
- Validación de `sha256` cuando exista y no sea `CHANGE_ME`.
- Si `sha256` es `CHANGE_ME` o vacío, se muestra `WARN` claro y se pide confirmación (o se permite con `--force` si se añade en el futuro; de momento se avisa y continúa, pero de forma explícita).
- Path traversal protection en extracción de `zip` y `tar.gz`: se rechazan entradas que contengan `..` o que salgan del directorio destino.
- Sanitización de nombres de archivo dentro de archives.
- No se sobrescriben binarios sin flujo controlado: `install` falla si ya existe, a menos que se trate de `update`.
- Timeouts HTTP: 10s para registry, 60s para descarga de assets.
- No se ejecutan scripts ni binarios descargados automáticamente (solo se extraen y se colocan).
- No telemetría, no envío de datos privados al backend.
- Errores humanos: sin stack traces al usuario.

## 12. Flujo de instalación

1. `cu install <tool>`
2. Resolver `tool` por id o name (case-insensitive) en registry remoto o caché.
3. Detectar plataforma actual (`runtime.GOOS` + `runtime.GOARCH` → `linux-amd64`, etc.).
4. Verificar que el asset exista para esa plataforma.
5. Descargar asset a un archivo temporal.
6. Si `sha256` existe y no es `CHANGE_ME`, validar. Si falla, abortar y borrar temporal.
7. Si `sha256` es `CHANGE_ME`, mostrar warning claro antes de continuar.
8. Crear directorio destino: `~/.core-utils/tools/<tool-id>/<version>/`
9. Extraer archive (zip/tar.gz) con protección anti path traversal.
10. Localizar `binaryPath` dentro del archive extraído.
11. Copiar/enlazar el binario a `~/.core-utils/bin/<binary>`.
12. En Linux/macOS, asegurar permisos ejecutables (`0755`).
13. Guardar metadata en `installed.json`.
14. Si `~/.core-utils/bin` no está en `PATH`, mostrar `WARN` con instrucciones.

## 13. Flujo de actualización

- `cu update` (sin args): iterar `installed.json`, consultar registry, comparar versiones semánticamente (string equality simple por ahora; si difiere, es nueva). Si hay nueva, repetir flujo de instalación y reemplazar.
- `cu update <tool>`: igual, pero solo para esa tool.
- Si el asset nuevo no existe para la plataforma actual, abortar con error claro.
- Si el binario anterior está en uso, se descarga primero a temp, se valida, se extrae a la nueva versión, se actualiza el symlink/binario en `bin/`, y se actualiza `installed.json`.

## 14. Flujo de eliminación

1. `cu remove <tool>`
2. Resolver en `installed.json`.
3. Borrar `~/.core-utils/tools/<id>/` (recursive).
4. Borrar `~/.core-utils/bin/<binary>`.
5. Quitar entrada de `installed.json`.
6. Si no se encuentra, devolver error claro.

## 15. Checklist de implementación

### Backend
- [ ] Crear `app/routers/registry.py`
- [ ] Crear `app/schemas/registry.py`
- [ ] Crear `app/models/registry.py` (definir, no activar todavía)
- [ ] Crear `data/registry.json` con datos iniciales estáticos
- [ ] Registrar router en `app/main.py`
- [ ] Asegurar que el backend arranca sin errores
- [ ] Verificar endpoints con curl/browser

### CLI
- [ ] Inicializar Go module (`go mod init`)
- [ ] Crear estructura de paquetes internos
- [ ] Implementar `config` (paths, env, JSON I/O)
- [ ] Implementar `registry` (fetch, parse, search)
- [ ] Implementar `platform` (detección, mapping)
- [ ] Implementar `checksum` (sha256)
- [ ] Implementar `archive` (zip + tar.gz, path traversal guard)
- [ ] Implementar `installer` (install, update, remove)
- [ ] Implementar `output` (pretty, status)
- [ ] Implementar `cli/commands.go` (todos los comandos)
- [ ] `cmd/cu/main.go` (entrypoint)
- [ ] `scripts/build.sh` y `scripts/install-local.sh`
- [ ] Tests unitarios
- [ ] `README.md`
- [ ] `go build -o bin/cu ./cmd/cu`
- [ ] `go test ./...`
- [ ] `gofmt`

## 16. Cómo probar backend y CLI

### Backend
```bash
cd core-utils-backend
source .venv/bin/activate
uvicorn app.main:app --reload

# Probar registry
curl http://localhost:8000/api/registry
curl http://localhost:8000/api/registry/tools
curl http://localhost:8000/api/registry/tools/memory-note-cli
curl http://localhost:8000/api/registry/tools/memory-note-cli/releases/latest
```

### CLI
```bash
cd core-utils-projects/core-utils-cli

# Build
chmod +x scripts/build.sh
./scripts/build.sh

# Test
CORE_UTILS_REGISTRY_URL=http://localhost:8000/api/registry ./bin/cu doctor
CORE_UTILS_REGISTRY_URL=http://localhost:8000/api/registry ./bin/cu registry sync
CORE_UTILS_REGISTRY_URL=http://localhost:8000/api/registry ./bin/cu search note
CORE_UTILS_REGISTRY_URL=http://localhost:8000/api/registry ./bin/cu info memory-note-cli
```

## 17. Riesgos técnicos y decisiones tomadas

| Riesgo | Decisión |
|--------|----------|
| Base de datos nueva rompe deploy o SQLite | Usar JSON estático local en `data/registry.json` como MVP. Las tablas SQLAlchemy se definen pero no se usan activamente todavía. |
| URLs falsas en registry | Todas las URLs sin release real se marcan como `TODO` y `sha256` como `CHANGE_ME`. |
| CLI depende de releases reales que no existen | El CLI funciona completo con datos de ejemplo; cuando haya releases reales solo hay que actualizar `data/registry.json` (o la DB). |
| Path traversal en extracción | Implementar guard manual en `archive.go`: rechazar `..`, normalizar paths, verificar que el archivo extraído quede dentro del directorio destino. |
| Compatibilidad de plataformas | Usar `runtime.GOOS` + `runtime.GOARCH` con un mapping simple: `linux/amd64` → `linux-amd64`. Extensible. |
| PATH no configurado | El `doctor` y el `install` advierten si `~/.core-utils/bin` no está en `PATH`. |
| No hay tests existentes en backend | No forzar un framework nuevo. Dejar validación manual documentada. Si el backend ya tuviera pytest, se añadirían tests, pero no es el caso. |
| Go version en el entorno | Asumir Go 1.22+ disponible. Si no lo está, documentar. |

---

**Fecha del plan**: 2026-06-11
**Autor**: OpenCode
