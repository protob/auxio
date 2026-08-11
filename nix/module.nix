self:
{ config, lib, pkgs, ... }:

let cfg = config.services.auxio;
in {
  options.services.auxio = {
    enable = lib.mkEnableOption "Auxio S3-compatible storage";

    package = lib.mkOption {
      type = lib.types.package;
      default = self.packages.${pkgs.system}.default;
      description = "The Auxio package to use.";
    };

    bind = lib.mkOption {
      type = lib.types.str;
      default = "127.0.0.1";
      description = ''
        Bind address. Loopback by default; put Auxio behind a reverse proxy (e.g.
        Caddy) rather than binding a public interface such as "0.0.0.0". On a
        non-loopback bind Auxio refuses to start while the S3 keys are still the
        built-in defaults.
      '';
    };

    port = lib.mkOption {
      type = lib.types.port;
      default = 9000;
      description = "HTTP listen port.";
    };

    dataDir = lib.mkOption {
      type = lib.types.path;
      default = "/var/lib/auxio";
      description = "Directory for bucket data and SQLite database.";
    };

    region = lib.mkOption {
      type = lib.types.str;
      default = "us-east-1";
      description = "S3 region string returned in responses.";
    };

    environmentFile = lib.mkOption {
      type = lib.types.nullOr lib.types.path;
      default = null;
      description = ''
        Path to an environment file containing AUXIO_ACCESS_KEY and
        AUXIO_SECRET_KEY, and - for any deployment reachable beyond loopback -
        AUXIO_USERNAME and AUXIO_PASSWORD for the dashboard. Leaving the
        dashboard credentials at the built-in admin/password is unsafe on an
        internet-facing deployment (Auxio logs a warning at startup when they
        are still the defaults). Use sops-nix or agenix to manage this file.
      '';
    };

    imaging = {
      enable = lib.mkEnableOption "image processing (requires libvips)";

      maxWidth = lib.mkOption {
        type = lib.types.int;
        default = 4096;
        description = "Maximum output image width.";
      };

      maxHeight = lib.mkOption {
        type = lib.types.int;
        default = 4096;
        description = "Maximum output image height.";
      };
    };

    uploadCleanupHours = lib.mkOption {
      type = lib.types.int;
      default = 24;
      description = "Hours before abandoned multipart uploads are cleaned up.";
    };
  };

  config = lib.mkIf cfg.enable {
    assertions = [{
      assertion = cfg.environmentFile != null;
      message =
        "services.auxio.environmentFile must be set (contains AUXIO_ACCESS_KEY and AUXIO_SECRET_KEY)";
    }];

    systemd.services.auxio = {
      description = "Auxio S3-Compatible Object Storage";
      after = [ "network.target" ];
      wantedBy = [ "multi-user.target" ];

      environment = {
        AUXIO_BIND = cfg.bind;
        AUXIO_HTTP_PORT = toString cfg.port;
        AUXIO_DATA_DIR = cfg.dataDir;
        AUXIO_REGION = cfg.region;
        AUXIO_IMAGING_ENABLED = lib.boolToString cfg.imaging.enable;
        AUXIO_IMAGING_MAX_WIDTH = toString cfg.imaging.maxWidth;
        AUXIO_IMAGING_MAX_HEIGHT = toString cfg.imaging.maxHeight;
        AUXIO_UPLOAD_CLEANUP_HOURS = toString cfg.uploadCleanupHours;
      };

      serviceConfig = {
        ExecStart = "${cfg.package}/bin/auxio";
        EnvironmentFile = cfg.environmentFile;
        DynamicUser = true;
        StateDirectory = "auxio";
        Restart = "always";
        RestartSec = 5;

        NoNewPrivileges = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        ReadWritePaths = [ cfg.dataDir ];
        PrivateTmp = true;
        PrivateDevices = true;
        ProtectKernelTunables = true;
        ProtectControlGroups = true;
        RestrictSUIDSGID = true;
      };
    };
  };
}
