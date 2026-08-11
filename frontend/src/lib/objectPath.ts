// S3 keys are absolute, but the table shows one directory level at a time, so
// every display name is the key minus the prefix currently being browsed.

export interface Crumb {
  label: string
  path: string
}

export function objectName(key: string, prefix: string): string {
  return key.slice(prefix.length)
}

export function folderName(folder: string, prefix: string): string {
  return folder.slice(prefix.length).replace(/\/$/, '')
}

export function fileUrl(bucket: string, key: string): string {
  return `/${bucket}/${key}`
}

// 52px for a 26px box: the image pipeline gets the 2x source, the browser
// downscales it.
export function thumbUrl(bucket: string, key: string): string {
  return `${fileUrl(bucket, key)}?w=52&h=52&fit=cover&fmt=webp`
}

// The bucket itself is the first crumb and its path is '' - the empty prefix is
// the bucket root, not a missing value.
export function breadcrumbs(bucket: string, prefix: string): Crumb[] {
  const crumbs: Crumb[] = [{ label: bucket, path: '' }]
  let path = ''
  for (const part of prefix.split('/').filter(Boolean)) {
    path += `${part}/`
    crumbs.push({ label: part, path })
  }
  return crumbs
}
