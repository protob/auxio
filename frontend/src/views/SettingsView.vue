<script setup lang="ts">
// VISUAL MOCKUP ONLY - no backend wiring. Auxio reads config from AUXIO_* env
// vars, so there is nothing to save; the values below are the built-in defaults.
// Swap them for live data and enable the controls once a config endpoint exists.
import { ref } from 'vue'
import SettingRow from '@/components/settings/SettingRow.vue'
import PrtBtn from '@/components/ui/btn/PrtBtn.vue'
import PrtInput from '@/components/ui/input/PrtInput.vue'
import PrtNotice from '@/components/ui/notice/PrtNotice.vue'
import PrtSegmented from '@/components/ui/segmented/PrtSegmented.vue'
import PrtTabs from '@/components/ui/tabs/PrtTabs.vue'
import PrtToggle from '@/components/ui/toggle/PrtToggle.vue'
import IconEye from '~icons/lucide/eye'
import IconEyeOff from '~icons/lucide/eye-off'
import IconTrash2 from '~icons/lucide/trash-2'

// PrtTabs' model is string | number, so the tab id stays a plain string and the
// panels come from its slots rather than a v-if chain.
const tab = ref('general')

const logLevel = ref('info')
const imagingEnabled = ref(true)
const showSecret = ref(false)

const tabs = [
  { value: 'general', label: 'General' },
  { value: 's3', label: 'S3 API' },
  { value: 'imaging', label: 'Image Processing' },
  { value: 'security', label: 'Security' },
]
const logLevels = ['debug', 'info', 'warn', 'error'].map(l => ({ value: l, label: l }))

// Complete literal strings, one per field width. `read-only:` carries a
// pseudo-class, so the muted ink outranks the variant's text-ink on specificity
// rather than racing it on emission order.
const fieldClass = 'w-70 font-mono read-only:text-ink-muted read-only:cursor-default'
const numFieldClass = 'w-22.5 text-right font-mono read-only:text-ink-muted read-only:cursor-default'
const secretFieldClass = 'w-59 font-mono read-only:text-ink-muted read-only:cursor-default'
</script>

<template>
  <div class="max-w-195">
    <h1 class="text-[1.5625rem] font-semibold text-ink tracking-[-0.05rem]">Settings</h1>
    <p class="mt-1.75 text-sm text-ink-muted leading-relaxed">
      Server configuration. Auxio reads these from
      <span class="font-mono text-[0.8125rem] text-ink">AUXIO_*</span> environment variables — usually your systemd unit.
    </p>

    <PrtNotice
      intent="info"
      variant="wash"
      class="mt-5"
      description="A read-only preview of the .env — the values below are the built-in defaults, not what this server is currently running."
    />

    <PrtTabs v-model="tab" :items="tabs" class="mt-6">
      <template #panel-general>
        <div class="divide-y divide-edge">
          <SettingRow
            label="Data directory"
            env="AUXIO_DATA_DIR"
            description="Root directory for buckets, the SQLite index, and image cache."
          >
            <PrtInput variant="outline" model-value="./data" readonly :class="fieldClass" />
          </SettingRow>
          <SettingRow
            label="Bind address"
            env="AUXIO_BIND"
            description="Interface to listen on. Loopback by default; Caddy fronts it for TLS."
          >
            <PrtInput variant="outline" model-value="127.0.0.1" readonly :class="fieldClass" />
          </SettingRow>
          <SettingRow
            label="HTTP port"
            env="AUXIO_HTTP_PORT"
            description="Port the S3 API and dashboard listen on."
          >
            <PrtInput variant="outline" model-value="9000" readonly :class="fieldClass" />
          </SettingRow>
          <SettingRow
            label="Region"
            env="AUXIO_REGION"
            description="S3 region identifier — for SDK compatibility only."
          >
            <PrtInput variant="outline" model-value="eu-north-1" readonly :class="fieldClass" />
          </SettingRow>
          <SettingRow
            label="Multipart cleanup"
            env="AUXIO_UPLOAD_CLEANUP_HOURS"
            description="Abandon incomplete multipart uploads after this many hours."
          >
            <PrtInput variant="outline" model-value="24" readonly :class="numFieldClass" />
            <span class="text-sm text-ink-muted">hours</span>
          </SettingRow>
          <SettingRow label="Log level" env="AUXIO_LOG_LEVEL" description="Verbosity of server logs.">
            <PrtSegmented v-model="logLevel" :options="logLevels" seed="7" size="sm" label="Log level" />
          </SettingRow>
        </div>
      </template>

      <template #panel-s3>
        <PrtNotice intent="info" variant="edge" class="mb-1">
          Endpoint <span class="font-mono text-ink">http://localhost:9000</span> ·
          Auth <span class="font-mono text-ink">AWS SigV4</span>
        </PrtNotice>
        <div class="divide-y divide-edge">
          <SettingRow
            label="Access key"
            env="AUXIO_ACCESS_KEY"
            description="S3 access key ID. Use a strong value in production."
          >
            <PrtInput variant="outline" model-value="auxio" readonly :class="fieldClass" />
          </SettingRow>
          <SettingRow
            label="Secret key"
            env="AUXIO_SECRET_KEY"
            description="S3 secret access key. Keep this private."
          >
            <PrtInput
              variant="outline"
              :type="showSecret ? 'text' : 'password'"
              model-value="auxio-secret-key"
              readonly
              :class="secretFieldClass"
            />
            <PrtBtn
              variant="outline"
              :title="showSecret ? 'Hide' : 'Reveal'"
              :aria-label="showSecret ? 'Hide secret key' : 'Reveal secret key'"
              :aria-pressed="showSecret"
              @click="showSecret = !showSecret"
            >
              <IconEyeOff v-if="showSecret" class="w-3.5 h-3.5" />
              <IconEye v-else class="w-3.5 h-3.5" />
            </PrtBtn>
          </SettingRow>
        </div>
      </template>

      <template #panel-imaging>
        <div class="divide-y divide-edge">
          <SettingRow
            label="Enable image processing"
            env="AUXIO_IMAGING_ENABLED"
            description="On-the-fly resize, crop, and format conversion via libvips."
          >
            <!-- No label prop: the row heading already says it, and PrtToggle
                 sends $attrs to the native input, so aria-label reaches it. -->
            <PrtToggle v-model="imagingEnabled" seed="7" aria-label="Enable image processing" />
          </SettingRow>
          <SettingRow
            label="Max dimensions"
            env="AUXIO_IMAGING_MAX_WIDTH / _MAX_HEIGHT"
            description="Upper bound for transformed output, in pixels."
            :dimmed="!imagingEnabled"
          >
            <PrtInput variant="outline" model-value="4096" readonly :class="numFieldClass" />
            <span class="text-ink-faint">×</span>
            <PrtInput variant="outline" model-value="4096" readonly :class="numFieldClass" />
          </SettingRow>
          <SettingRow label="Per-request transforms" :dimmed="!imagingEnabled">
            <template #description>
              Quality and format are chosen per request via query params —
              <span class="font-mono">?w=</span> <span class="font-mono">?h=</span>
              <span class="font-mono">?fit=</span> <span class="font-mono">?q=</span>
              <span class="font-mono">?fmt=</span> — not a server-wide default.
            </template>
          </SettingRow>
          <SettingRow label="Image cache">
            <template #description><span class="font-mono">{DATA_DIR}/.imgcache/</span></template>
            <!-- aria-disabled, not disabled: the reason is in the title, and a
                 disabled button cannot be focused to read it. -->
            <PrtBtn variant="outline" size="sm" aria-disabled="true" title="Not wired up yet">
              <IconTrash2 class="w-3.5 h-3.5" /> Clear cache
            </PrtBtn>
          </SettingRow>
        </div>
      </template>

      <template #panel-security>
        <!-- Conditional wording, not a warning: the page shows the built-in
             defaults, it cannot see what the server was actually started with. -->
        <PrtNotice intent="warning" variant="wash" class="mb-1">
          If these are still the values below, set
          <span class="font-mono">AUXIO_USERNAME</span> and <span class="font-mono">AUXIO_PASSWORD</span>
          before exposing this server.
        </PrtNotice>
        <div class="divide-y divide-edge">
          <SettingRow
            label="Dashboard username"
            env="AUXIO_USERNAME"
            description="Login for this web dashboard."
          >
            <PrtInput variant="outline" model-value="admin" readonly :class="fieldClass" />
          </SettingRow>
          <SettingRow
            label="Dashboard password"
            env="AUXIO_PASSWORD"
            description="Authenticated sessions use a Bearer token."
          >
            <PrtInput variant="outline" type="password" model-value="password" readonly :class="fieldClass" />
          </SettingRow>
        </div>
      </template>
    </PrtTabs>

    <div class="flex items-center justify-end gap-2.5 mt-8 pt-5.5 border-t border-edge">
      <span class="mr-auto text-xs text-ink-faint">Configured via environment variables</span>
      <PrtBtn variant="outline" aria-disabled="true" title="Not wired up yet">Reset to defaults</PrtBtn>
      <PrtBtn seed="7" aria-disabled="true" title="Not wired up yet">Save &amp; restart</PrtBtn>
    </div>
  </div>
</template>
