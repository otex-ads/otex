export const TAG_URL = "https://cdn.otexads.com/tag.js";

export interface AdCodeOptions {
  zoneId: string;
  async?: boolean;
  responsive?: boolean;
  customStyle?: string;
}

export function buildAdCode({ zoneId, async = true }: AdCodeOptions) {
  return `<!-- OtexAds zone ${zoneId} -->
<script${async ? " async" : ""} src="${TAG_URL}" data-zone-id="${zoneId}"></script>
`;
}
