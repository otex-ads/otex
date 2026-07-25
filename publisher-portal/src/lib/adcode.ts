import { API_BASE } from "./api";

export interface AdCodeOptions {
  zoneId: string;
  async?: boolean;
  responsive?: boolean;
  customStyle?: string;
}

export function buildAdCode({
  zoneId,
  async = true,
  responsive = true,
  customStyle,
}: AdCodeOptions) {
  const style =
    customStyle?.trim() || (responsive ? "max-width:100%;margin:0 auto;display:block;" : "");
  return `<!-- PropelAds zone ${zoneId} -->
<div id="pa-zone-${zoneId}" style="${style}"></div>
<script${async ? " async" : ""} src="${API_BASE}/ads/loader.js?zone=${zoneId}"></script>
`;
}
