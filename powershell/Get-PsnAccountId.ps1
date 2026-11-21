<#
.SYNOPSIS
    Get your PSN Account ID for PS5 Remote Play using only Sony's own login.

.DESCRIPTION
    Talks only to auth.api.sonyentertainmentnetwork.com over HTTPS. Saves nothing.
    Works with the Windows PowerShell 5.1 built into Windows 10/11 and PowerShell 7+.
    Read it before running it - that is the point of this version.

.EXAMPLE
    powershell -ExecutionPolicy Bypass -File .\Get-PsnAccountId.ps1
.EXAMPLE
    .\Get-PsnAccountId.ps1 -Convert 1234567890123456789
.EXAMPLE
    .\Get-PsnAccountId.ps1 -Lang es      # en español
#>
[CmdletBinding()]
param(
    [switch]$Json,
    [switch]$NoBrowser,
    [string]$Convert,
    [ValidateSet('en', 'es')][string]$Lang
)

$ErrorActionPreference = 'Stop'
$AuthHost = 'https://auth.api.sonyentertainmentnetwork.com'
# Public identifiers of Sony's PS Remote Play app, as used by chiaki / chiaki-ng.
# They identify the app, not you, and grant nothing without your own Sony login.
$ClientId = 'ba495a24-818c-472b-b12d-ff231c1b5745'
$ClientSecret = 'mvaiZkRsAsI1IBkY'
$Redirect = 'https://remoteplay.dl.playstation.net/remoteplay/redirect'
$Scope = 'psn:clientapp referenceDataService:countryConfig.read pushNotification:webSocket.desktop.connect sessionManager:remotePlaySession.system.update'

# Language: -Lang wins, otherwise the Windows display language.
if (-not $Lang) { $Lang = if ((Get-UICulture).TwoLetterISOLanguageName -eq 'es') { 'es' } else { 'en' } }
$Messages = @{   # key = @(English, Spanish)
    step1a = @("Step 1. Sign in on Sony's page. Check the address bar shows a sony.com or", 'Paso 1. Inicia sesión en la página de Sony. Comprueba que la barra de direcciones muestra')
    step1b = @('        sonyentertainmentnetwork.com address before typing your password.', '        sony.com o sonyentertainmentnetwork.com antes de escribir tu contraseña.')
    step2a = @('Step 2. After signing in you land on a page starting with', 'Paso 2. Después de iniciar sesión llegarás a una página que empieza por')
    step2b = @("        (it may look blank or show an error - that's normal).", '        (puede verse en blanco o mostrar un error: es normal).')
    step2c = @('        Copy the WHOLE address from the address bar and paste it below.', '        Copia la dirección COMPLETA de la barra de direcciones y pégala abajo.')
    paste = @('Paste here', 'Pega aquí')
    result = @('Your PSN Account ID (paste this into Veyra or chiaki-ng):', 'Tu Account ID de PSN (pégalo en Veyra o en chiaki-ng):')
    number = @('Number: ', 'Número: ')
    account = @('Account:', 'Cuenta: ')
    check_you = @('(check this is you)', '(comprueba que eres tú)')
    footer = @('This ID never changes - you only need it once. Nothing was saved.', 'Este ID nunca cambia: solo lo necesitas una vez. No se ha guardado nada.')
    nothing_pasted = @('Nothing was pasted.', 'No se ha pegado nada.')
    not_redirect = @("That address is not Sony's redirect page. It must start with {0}?code=", 'Esa dirección no es la página de redirección de Sony. Debe empezar por {0}?code=')
    no_code = @("No 'code=' found in that address. Sign in again and copy the whole address.", "No hay 'code=' en esa dirección. Vuelve a iniciar sesión y copia la dirección completa.")
    invalid_code = @("That doesn't look like a valid code.", 'Eso no parece un código válido.')
    too_large = @('Number is too large to be a PSN Account ID.', 'El número es demasiado grande para ser un Account ID de PSN.')
    not_b64 = @('Not a number and not valid Base64.', 'No es un número ni Base64 válido.')
    wrong_len = @('A Base64 Account ID must decode to exactly 8 bytes.', 'Un Account ID en Base64 debe decodificarse en exactamente 8 bytes.')
    rejected = @('Sony rejected the code. Codes work once and expire within minutes: run the script again and paste faster.', 'Sony ha rechazado el código. Cada código sirve una sola vez y caduca en pocos minutos: vuelve a ejecutar el script y pega más rápido.')
    http_status = @('Sony returned HTTP {0}. Try again later.', 'Sony ha devuelto HTTP {0}. Inténtalo más tarde.')
    unreachable = @('Could not reach Sony ({0}). Check your internet connection.', 'No se ha podido conectar con Sony ({0}). Comprueba tu conexión a internet.')
    no_token = @('Sony did not return an access token.', 'Sony no ha devuelto un token de acceso.')
    no_number = @('Sony did not return a numeric account number.', 'Sony no ha devuelto un número de cuenta.')
}
function T([string]$Key, [object[]]$Arguments = @()) {
    $text = $Messages[$Key][$(if ($Lang -eq 'es') { 1 } else { 0 })]
    if ($Arguments.Count) { return ($text -f $Arguments) }
    return $text
}

function ConvertTo-AccountId([string]$Value) {
    $Value = $Value.Trim()
    if ($Value -match '^\d+$') {
        [UInt64]$n = 0
        if (-not [UInt64]::TryParse($Value, [ref]$n)) { throw (T 'too_large') }
        $bytes = [BitConverter]::GetBytes($n)
        if (-not [BitConverter]::IsLittleEndian) { [Array]::Reverse($bytes) }
        return [pscustomobject]@{ account_id_base64 = [Convert]::ToBase64String($bytes); account_id_number = $Value }
    }
    try { $raw = [Convert]::FromBase64String($Value) } catch { throw (T 'not_b64') }
    if ($raw.Length -ne 8) { throw (T 'wrong_len') }
    if (-not [BitConverter]::IsLittleEndian) { [Array]::Reverse($raw) }
    return [pscustomobject]@{ account_id_base64 = $Value; account_id_number = [BitConverter]::ToUInt64($raw, 0).ToString() }
}

function Get-LoginUrl {
    $q = [ordered]@{
        service_entity = 'urn:service-entity:psn'; response_type = 'code'; client_id = $ClientId
        redirect_uri = $Redirect; scope = $Scope; request_locale = $(if ($Lang -eq 'es') { 'es_ES' } else { 'en_US' }); ui = 'pr'
        service_logo = 'ps'; layout_type = 'popup'; smcid = 'remoteplay'; prompt = 'always'
        PlatformPrivacyWs1 = 'minimal'
    }
    $parts = foreach ($k in $q.Keys) { "$k=$([Uri]::EscapeDataString($q[$k]))" }
    return "$AuthHost/2.0/oauth/authorize?" + ($parts -join '&')
}

function Get-CodeFromPaste([string]$Pasted) {
    $text = $Pasted.Trim().Trim('"', "'")
    if (-not $text) { throw (T 'nothing_pasted') }
    if ($text -match '^https?://') {
        if (-not $text.StartsWith("${Redirect}?")) { throw (T 'not_redirect' @($Redirect)) }
        $codes = @(($text.Substring($Redirect.Length + 1) -split '&') | Where-Object { $_ -like 'code=*' })
        if ($codes.Count -ne 1 -or $codes[0].Length -le 5) { throw (T 'no_code') }
        $text = [Uri]::UnescapeDataString($codes[0].Substring(5))
    }
    if ($text -match '\s' -or $text.Length -gt 2048) { throw (T 'invalid_code') }
    return $text
}

function Invoke-Sony([string]$Method, [string]$Url, $Body) {
    $basic = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("${ClientId}:${ClientSecret}"))
    $params = @{ Method = $Method; Uri = $Url; Headers = @{ Authorization = "Basic $basic" }; UseBasicParsing = $true; TimeoutSec = 30 }
    if ($Body) { $params.Body = $Body; $params.ContentType = 'application/x-www-form-urlencoded' }
    try { return Invoke-RestMethod @params }
    catch {
        $status = 0
        if ($_.Exception.Response) { $status = [int]$_.Exception.Response.StatusCode }
        if ($status -eq 400 -or $status -eq 401) { throw (T 'rejected') }
        if ($status) { throw (T 'http_status' @($status)) }
        throw (T 'unreachable' @($_.Exception.Message))
    }
}

function Get-PsnAccount([string]$Code) {
    $form = "grant_type=authorization_code&code=$([Uri]::EscapeDataString($Code))&scope=$([Uri]::EscapeDataString($Scope))&redirect_uri=$([Uri]::EscapeDataString($Redirect))"
    $token = Invoke-Sony 'POST' "$AuthHost/2.0/oauth/token" $form
    if (-not $token.access_token) { throw (T 'no_token') }
    $info = Invoke-Sony 'GET' "$AuthHost/2.0/oauth/token/$([Uri]::EscapeDataString($token.access_token))" $null
    if ("$($info.user_id)" -notmatch '^\d+$') { throw (T 'no_number') }
    $result = ConvertTo-AccountId "$($info.user_id)"
    $result | Add-Member -NotePropertyName online_id -NotePropertyValue "$($info.online_id)"
    return $result
}

# Windows PowerShell 5.1 may not enable TLS 1.2 on older systems; add it without removing others.
[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

try {
    if ($Convert) {
        $result = ConvertTo-AccountId $Convert
    } else {
        $url = Get-LoginUrl
        Write-Host (T 'step1a')
        Write-Host ((T 'step1b') + "`n")
        Write-Host "   $url`n"
        if (-not $NoBrowser) { Start-Process $url }
        Write-Host (T 'step2a')
        Write-Host "        ${Redirect}?code=..."
        Write-Host (T 'step2b')
        Write-Host ((T 'step2c') + "`n")
        $result = Get-PsnAccount (Get-CodeFromPaste (Read-Host (T 'paste')))
    }
} catch {
    Write-Host "`nError: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

if ($Json) { $result | ConvertTo-Json; exit 0 }
Write-Host ("`n" + (T 'result') + "`n")
Write-Host "   Base64:  $($result.account_id_base64)" -ForegroundColor Green
Write-Host "   $(T 'number') $($result.account_id_number)"
if ($result.online_id) { Write-Host "   $(T 'account') $($result.online_id)  $(T 'check_you')" }
Write-Host ("`n" + (T 'footer'))
