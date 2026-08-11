export const TOKEN_KEY = 'auxio-token'
// Not 'auxio:groupBy' - that key still holds 'on'/'off' strings in the wild,
// which useLocalStorage would read back as false.
export const GROUP_BY_KEY = 'auxio:group-by'
export const COLLAPSED_KEY = 'auxio:collapsedGroups'
// The sidebar's icon rail. A boolean, so it cannot share COLLAPSED_KEY, which
// holds the group-collapse array for the bucket table.
export const SIDEBAR_RAILED_KEY = 'auxio:sidebar-railed'

export const BUCKET_PAGE_SIZE = 12
export const OBJECT_MAX_KEYS = 1000

// Browsers cap connections per host anyway; the point is to stop 300 dropped
// files from each holding a FormData buffer while they queue.
export const UPLOAD_CONCURRENCY = 4
