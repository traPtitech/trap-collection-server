const validChars =
  'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'

export function randomString(length) {
  let array = new Uint8Array(length)
  window.crypto.getRandomValues(array)
  array = array.map(x => validChars.charCodeAt(x % validChars.length))
  return String.fromCharCode(...array)
}

const b64Chars = { '+': '-', '/': '_', '=': '' }

function urlEncodeB64(input) {
  return input.replace(/[+/=]/g, m => b64Chars[m])
}

function bufferToBase64UrlEncoded(input) {
  const bytes = new Uint8Array(input)
  return urlEncodeB64(window.btoa(String.fromCharCode(...bytes)))
}

function sha256(message) {
  const data = new TextEncoder().encode(message)
  return window.crypto.subtle.digest('SHA-256', data)
}

function buf2hex(buf) {
  return Array.prototype.map
    .call(new Uint8Array(buf), x => ('00' + x.toString(16)).slice(-2))
    .join('')
}

export async function hmacsha1(message, secret) {
  const key = await window.crypto.subtle.importKey(
    'raw',
    new TextEncoder().encode(secret),
    {
      name: 'HMAC',
      hash: 'SHA-1'
    },
    true,
    ['sign']
  )
  const buf = new TextEncoder().encode(message)
  const sig = await window.crypto.subtle.sign('HMAC', key, buf)
  return buf2hex(sig)
}

export async function pkce(verifier) {
  return sha256(verifier).then(bufferToBase64UrlEncoded)
}
