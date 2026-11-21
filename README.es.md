# psn-account-id

[English](README.md) | **Español**

**Obtén tu Account ID de PlayStation Network de forma segura, usando solo la página de inicio de sesión de Sony, para poder retransmitir tu PS5 a tu PC.**

Sin webs de terceros, sin crear cuentas, sin guardar nada, y con el código abierto para que cualquiera pueda leerlo.

---

## En pocas palabras

### ¿Qué es esto?
Una pequeña herramienta gratuita. Inicias sesión en **la propia web de Sony** y la herramienta te muestra tu **Account ID de PSN**, un número que identifica tu cuenta de PlayStation.

### ¿Para qué necesito mi Account ID?
Los programas que retransmiten tu PS5 a tu ordenador (el "Juego remoto" de PS5) te lo piden una vez, al emparejar el programa con tu consola. Por ejemplo **[Veyra](https://github.com/Likely7/Veyra-NRVideo)**, que se usa para jugar a GTA 6 con el escalado por IA de NVIDIA, y **[chiaki-ng](https://github.com/chiaki-ng/chiaki-ng)**. La PS5 solo acepta el emparejamiento si el Account ID corresponde a un usuario de esa consola **y** además escribes el código de 8 cifras que la PS5 muestra en pantalla.

El Account ID **no es una contraseña**. Por sí solo no da acceso a tu cuenta ni a tu consola. Solo lo necesitas **una vez**: no cambia nunca, ni aunque cambies tu contraseña o tu nombre de PSN.

### ¿Por qué no usar la web que enlaza Veyra?
El botón **"Get Account ID"** (Obtener Account ID) de Veyra abre **psntools.com**, una web de terceros ([código fuente](https://github.com/Likely7/Veyra-NRVideo/blob/fe8df5328a62849eb2d79053564b9574ab68aecd/src/ui/bridge/RemotePlay.cpp#L515-L522)). El propio código de Veyra no envía nada a esa web; solo abre la página. Esto es lo que es esa web, comprobado el 21-11-2026:

- **Te pide tu cookie de inicio de sesión NPSSO para registrarte.** Su registro dice *"Unable to register without the NPSSO cookie"* ("No se puede registrar sin la cookie NPSSO") y te pide *"Paste your NPSSO cookie here"* ("Pega aquí tu cookie NPSSO"). **La cookie NPSSO funciona como tu contraseña**: quien la tenga puede actuar como tú en PlayStation Network. Su propia documentación dice que da acceso a tu *historial de transacciones, nombre real, fecha de nacimiento* y más, y su política de privacidad dice que **guarda** una copia cifrada.
- **Nadie sabe quién la gestiona.** El dueño del dominio está oculto en el registro público (WHOIS "DATA REDACTED") y no figura ninguna empresa.
- **También vende cuentas de PlayStation.** El código de la web contiene listados de venta de cuentas PSN ("raras" y "nuevas") y enlaces a valoraciones de vendedores en mercados de compraventa de cuentas de juegos.

**No** decimos que esa web robe cuentas. Decimos que **nunca deberías entregar tu cookie de inicio de sesión a una web anónima** solo para leer un número que Sony te muestra directamente. Esta herramienta hace precisamente eso: es Sony quien te muestra el número.

> **Si ya has pegado tu cookie NPSSO en alguna web:** entra en los ajustes de tu cuenta de PlayStation, elige **Cerrar sesión en todos los dispositivos** y cambia tu contraseña. Así la cookie antigua deja de servir. Activa también la **verificación en dos pasos**.

---

## Cómo usarlo (Windows, sin instalar nada)

1. Ve a la **[página de versiones (Releases)](../../releases/latest)** y descarga **`psn-account-id-windows-amd64.exe`**. Usa `windows-arm64` solo en portátiles ARM, como los Snapdragon.
2. Haz doble clic. Puede aparecer **"Windows protegió su PC"**, porque el programa es nuevo y no está firmado con un certificado de pago. Pulsa **Más información → Ejecutar de todas formas**. (¿Dudas? Usa la versión de PowerShell de abajo, donde puedes leer cada línea antes de ejecutarla, o comprueba la descarga como se explica en "¿Es seguro?".)
3. Se abrirá tu navegador en la **página de inicio de sesión de Sony**. Comprueba que la barra de direcciones muestra **`my.account.sony.com`** o **`sonyentertainmentnetwork.com`** e inicia sesión como siempre.
4. Después de iniciar sesión llegarás a una página que empieza por `https://remoteplay.dl.playstation.net/remoteplay/redirect?code=…`. **Puede verse en blanco o mostrar un error; es normal.** Haz clic en la barra de direcciones, selecciona todo y cópialo.
5. Pégalo en la ventana de la herramienta (en la consola de Windows se pega con clic derecho) y pulsa Intro. **Hazlo en uno o dos minutos**, porque el código caduca enseguida.
6. La herramienta te muestra tu Account ID. Copia la línea **Base64**, por ejemplo `FYHpffQQIhE=`.

La herramienta se muestra en español automáticamente si tu Windows está en español. Puedes forzarlo con `--lang es` (o `--lang en` para inglés).

### Después, en Veyra
Abre el panel de PS5 / Juego remoto. **No pulses "Get Account ID"**; pega tu ID en el campo Account ID. Luego, en la PS5, ve a **Ajustes → Sistema → Juego remoto → Vincular dispositivo** y escribe en Veyra el código de 8 cifras que aparece. Veyra recordará la consola a partir de entonces.

### O en chiaki-ng
chiaki-ng tiene su propio inicio de sesión de Sony integrado y puede obtener el ID por sí mismo. Si lo prefieres, pega el ID de esta herramienta en su pantalla de registro.

---

## Tres versiones, el mismo comportamiento

| Versión | Ideal para | Cómo ejecutarla |
|---|---|---|
| **Programa para Windows / macOS / Linux** (`go/`) | La mayoría de la gente | Descárgalo desde [Releases](../../releases/latest) y haz doble clic (en macOS/Linux, ejecútalo en un terminal) |
| **Script de PowerShell** (`powershell/Get-PsnAccountId.ps1`) | Usuarios de Windows que quieren **leer el código antes de ejecutarlo**; usa el PowerShell que ya viene con Windows 10/11 | Ver abajo |
| **Script de Python** (`python/psn_account_id.py`) | Quien ya tenga Python 3.9 o superior | `python psn_account_id.py` |

### Ejecutar la versión de PowerShell
Windows bloquea por defecto los scripts descargados. Ábrelo primero en el Bloc de notas y léelo; después abre **PowerShell** en la carpeta donde lo guardaste y ejecuta:

```powershell
powershell -ExecutionPolicy Bypass -File .\Get-PsnAccountId.ps1 -Lang es
```

`-ExecutionPolicy Bypass` solo afecta a esta ejecución y no cambia la configuración de tu sistema. **Consejo de seguridad general:** nunca pegues en PowerShell comandos sacados de internet sin entender qué hacen. Las estafas usan justo ese truco.

### Opciones (todas las versiones)
| Opción | Qué hace |
|---|---|
| `--json` / `-Json` | Muestra el resultado en JSON, para scripts |
| `--no-browser` / `-NoBrowser` | Muestra el enlace de Sony en lugar de abrir el navegador |
| `--convert <id>` / `-Convert <id>` | Convierte un Account ID entre su forma numérica y Base64, sin conexión |
| `--lang es` / `-Lang es` | Idioma: `es` (español) o `en` (inglés). Por defecto, el idioma del sistema |

### ¿Puedo guardar el resultado en un archivo?
La herramienta no guarda nada por sí misma, a propósito. Si quieres guardarlo: `psn-account-id --json > mi-id-psn.json` o, en PowerShell, `.\Get-PsnAccountId.ps1 -Json | Out-File mi-id-psn.json`. Guardar el Account ID no es peligroso: no es un secreto, y la herramienta nunca muestra ni guarda el token de inicio de sesión.

### ¿Por qué Base64?
El número y el texto Base64 son **el mismo ID escrito de dos formas**. Sony devuelve tu cuenta como un número, y la PS5 lo usa empaquetado en 8 bytes. Base64 es la forma estándar de escribir esos bytes como texto. **chiaki-ng** solo acepta Base64; **Veyra** acepta cualquiera de las dos. Por eso la herramienta muestra ambas, con Base64 primero.

---

## ¿Es seguro? (detalles)

- **Solo habla con Sony.** El único servidor al que se conecta es `auth.api.sonyentertainmentnetwork.com`, por HTTPS. Tu contraseña la escribes en la página de Sony, en tu navegador; nunca en esta herramienta.
- **No guarda nada.** No escribe archivos ni conserva tokens. El token de acceso, de corta duración, se usa para dos peticiones y luego se descarta y caduca solo.
- **Pide permisos limitados:** acceso básico de la aplicación, ajustes del país, notificaciones y sesiones de Juego remoto. Es el mismo conjunto que pide la aplicación oficial de Juego remoto de Sony. No tiene acceso a compras, monedero ni ajustes de la cuenta.
- **No usa bibliotecas de terceros.** Las versiones en Go, Python y PowerShell usan solo la biblioteca estándar.
- **Las versiones se compilan en público.** Cada archivo publicado lo compila GitHub Actions a partir del código de este repositorio ([`release.yml`](.github/workflows/release.yml)), va acompañado de `SHA256SUMS.txt` y lleva una **atestación de procedencia firmada**. Puedes verificar una descarga con:
  ```
  gh attestation verify psn-account-id-windows-amd64.exe --repo metalito/psn-account-id
  ```

### ¿Qué son el "client ID" y el "client secret" del código?
Identifican a la **aplicación** ante Sony, no a ti. Pertenecen a la aplicación oficial **PS Remote Play** de Sony, los publicó hace años el proyecto de código abierto [Chiaki](https://github.com/chiaki-ng/chiaki-ng), y también los usan chiaki-ng y Veyra. Las aplicaciones de escritorio no pueden mantener en secreto esos valores, así que por sí solos no abren nada sin tu propio inicio de sesión en Sony. Sony podría dejar de aceptarlos en cualquier momento; si eso ocurre, esta herramienta (y el inicio de sesión de chiaki-ng) necesitarían una actualización.

## Cómo funciona (técnico)
1. Abre la página OAuth de Sony: `GET /2.0/oauth/authorize?response_type=code&client_id=…&redirect_uri=https://remoteplay.dl.playstation.net/remoteplay/redirect&scope=…`
2. Inicias sesión. Sony redirige a `…/remoteplay/redirect?code=<código de un solo uso>` y tú pegas esa dirección.
3. `POST /2.0/oauth/token` con `grant_type=authorization_code`, autenticado por HTTP Basic con el client ID y el secret, devuelve un token de acceso.
4. `GET /2.0/oauth/token/<token de acceso>` devuelve `user_id` y `online_id`.
5. Account ID = `user_id` como entero de 64 bits sin signo en little-endian, codificado en Base64. Es la misma conversión que usa chiaki-ng.

## Compilarlo y probarlo tú mismo
```bash
cd go && go test ./... && go build .               # Go 1.24+
cd python && python -m unittest -v                  # Python 3.9+
pwsh -File powershell/Get-PsnAccountId.ps1 -Convert 1234567890123456789 -Lang es
```
La integración continua ([`ci.yml`](.github/workflows/ci.yml)) ejecuta las pruebas de Go y Python y prueba el script de PowerShell en **Windows PowerShell 5.1** real, en inglés y en español. Al publicar una etiqueta como `v1.0.0` se compila y publica una versión.

## Aviso legal
No está afiliado ni respaldado por Sony Interactive Entertainment. "PlayStation", "PS5" y "PSN" son marcas registradas de Sony. Igual que chiaki-ng, esta herramienta inicia sesión como la aplicación de Juego remoto de Sony, algo que quizá no permitan los términos de servicio de Sony; úsala bajo tu propio criterio. Se creó como alternativa más segura a las webs de terceros. Licencia [MIT](LICENSE).
