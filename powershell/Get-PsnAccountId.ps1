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
#>
[CmdletBinding()]
param(
    [switch]$Json,
    [switch]$NoBrowser,
    [string]$Convert
)

$ErrorActionPreference = 'Stop'
$AuthHost = 'https://auth.api.sonyentertainmentnetwork.com'
# Public identifiers of Sony's PS Remote Play app, as used by chiaki / chiaki-ng.
# They identify the app, not you, and grant nothing without your own Sony login.
$ClientId = 'ba495a24-818c-472b-b12d-ff231c1b5745'
$ClientSecret = 'mvaiZkRsAsI1IBkY'
$Redirect = 'https://remoteplay.dl.playstation.net/remoteplay/redirect'
$Scope = 'psn:clientapp referenceDataService:countryConfig.read pushNotification:webSocket.desktop.connect sessionManager:remotePlaySession.system.update'

function ConvertTo-AccountId([string]$Value) {
    $Value = $Value.Trim()
    if ($Value -match '^\d+$') {
        [UInt64]$n = 0
        if (-not [UInt64]::TryParse($Value, [ref]$n)) { throw 'Number is too large to be a PSN Account ID.' }
        $bytes = [BitConverter]::GetBytes($n)
        if (-not [BitConverter]::IsLittleEndian) { [Array]::Reverse($bytes) }
        return [pscustomobject]@{ account_id_base64 = [Convert]::ToBase64String($bytes); account_id_number = $Value }
    }
    try { $raw = [Convert]::FromBase64String($Value) } catch { throw 'Not a number and not valid Base64.' }
    if ($raw.Length -ne 8) { throw 'A Base64 Account ID must decode to exactly 8 bytes.' }
    if (-not [BitConverter]::IsLittleEndian) { [Array]::Reverse($raw) }
    return [pscustomobject]@{ account_id_base64 = $Value; account_id_number = [BitConverter]::ToUInt64($raw, 0).ToString() }
}

function Get-LoginUrl {
    $q = [ordered]@{
        service_entity = 'urn:service-entity:psn'; response_type = 'code'; client_id = $ClientId
        redirect_uri = $Redirect; scope = $Scope; request_locale = 'en_US'; ui = 'pr'
        service_logo = 'ps'; layout_type = 'popup'; smcid = 'remoteplay'; prompt = 'always'
        PlatformPrivacyWs1 = 'minimal'
    }
    $parts = foreach ($k in $q.Keys) { "$k=$([Uri]::EscapeDataString($q[$k]))" }
    return "$AuthHost/2.0/oauth/authorize?" + ($parts -join '&')
}

function Get-CodeFromPaste([string]$Pasted) {
    $text = $Pasted.Trim().Trim('"', "'")
    if (-not $text) { throw 'Nothing was pasted.' }
    if ($text -match '^https?://') {
        if (-not $text.StartsWith("${Redirect}?")) { throw "That address is not Sony's redirect page. It must start with ${Redirect}?code=" }
        $codes = @(($text.Substring($Redirect.Length + 1) -split '&') | Where-Object { $_ -like 'code=*' })
        if ($codes.Count -ne 1 -or $codes[0].Length -le 5) { throw "No 'code=' found in that address. Sign in again and copy the whole address." }
        $text = [Uri]::UnescapeDataString($codes[0].Substring(5))
    }
    if ($text -match '\s' -or $text.Length -gt 2048) { throw "That doesn't look like a valid code." }
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
        if ($status -eq 400 -or $status -eq 401) { throw 'Sony rejected the code. Codes work once and expire within minutes: run the script again and paste faster.' }
        if ($status) { throw "Sony returned HTTP $status. Try again later." }
        throw "Could not reach Sony ($($_.Exception.Message)). Check your internet connection."
    }
}

function Get-PsnAccount([string]$Code) {
    $form = "grant_type=authorization_code&code=$([Uri]::EscapeDataString($Code))&scope=$([Uri]::EscapeDataString($Scope))&redirect_uri=$([Uri]::EscapeDataString($Redirect))"
    $token = Invoke-Sony 'POST' "$AuthHost/2.0/oauth/token" $form
    if (-not $token.access_token) { throw 'Sony did not return an access token.' }
    $info = Invoke-Sony 'GET' "$AuthHost/2.0/oauth/token/$([Uri]::EscapeDataString($token.access_token))" $null
    if ("$($info.user_id)" -notmatch '^\d+$') { throw 'Sony did not return a numeric account number.' }
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
        Write-Host "Step 1. Sign in on Sony's page. Check the address bar shows a sony.com or"
        Write-Host "        sonyentertainmentnetwork.com address before typing your password.`n"
        Write-Host "   $url`n"
        if (-not $NoBrowser) { Start-Process $url }
        Write-Host "Step 2. After signing in you land on a page starting with"
        Write-Host "        ${Redirect}?code=..."
        Write-Host "        (it may look blank or show an error - that's normal)."
        Write-Host "        Copy the WHOLE address from the address bar and paste it below.`n"
        $result = Get-PsnAccount (Get-CodeFromPaste (Read-Host 'Paste here'))
    }
} catch {
    Write-Host "`nError: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

if ($Json) { $result | ConvertTo-Json; exit 0 }
Write-Host "`nYour PSN Account ID (paste this into Veyra or chiaki-ng):`n"
Write-Host "   Base64:  $($result.account_id_base64)" -ForegroundColor Green
Write-Host "   Number:  $($result.account_id_number)"
if ($result.online_id) { Write-Host "   Account: $($result.online_id)  (check this is you)" }
Write-Host "`nThis ID never changes - you only need it once. Nothing was saved."
