# Example: Caddy reverse proxy in front of Auxio
#
# Reference only - add this alongside services.auxio in your NixOS configuration.
# Auxio binds loopback (services.auxio.bind = "127.0.0.1") and Caddy terminates
# TLS in front of it. There is no Auxio "mode" flag: whether the deployment is
# internet-facing is decided by the cert source and domain below.

{
  services.caddy = {
    enable = true;

    # Workstation: Caddy's internal CA, resolves to loopback/LAN only.
    virtualHosts."auxio.app.localdev" = {
      extraConfig = ''
        tls internal
        reverse_proxy localhost:9000
      '';
    };

    # Server: real ACME cert, internet-facing.
    # virtualHosts."media.example.com" = {
    #   extraConfig = ''
    #     reverse_proxy localhost:9000
    #   '';
    # };
  };

  # For the workstation host, point the dev domain at loopback:
  # networking.extraHosts = ''
  #   127.0.0.1 auxio.app.localdev
  # '';
}

# Notes:
# - reverse_proxy port must match services.auxio.port (default 9000).
# - Caddy preserves the Host header by default and Auxio signs SigV4 over
#   r.Host, so no `header_up Host` or X-Forwarded-Host plumbing is needed.
# - Caddy is a dumb TLS terminator: no auth, no rewriting. All S3 auth (SigV4)
#   and the dashboard bearer token are handled by Auxio.
# - Optional operator-side lock: on an internet-facing server you can gate
#   /dashboard* and /api/* with Caddy `basic_auth` or `forward_auth`. S3
#   consumers only hit /{bucket}/{key}, so this is invisible to them.
