// OtexAds Ad Tag
// This script loads and displays ads for the specified zone

(function() {
  'use strict';

  // Find the script tag that loaded this file
  var scripts = document.getElementsByTagName('script');
  var currentScript = scripts[scripts.length - 1];
  var zoneId = currentScript.getAttribute('data-zone-id');

  if (!zoneId) {
    console.error('[OtexAds] Missing data-zone-id attribute on script tag');
    return;
  }

  // Create container for the ad
  var container = document.createElement('div');
  container.id = 'otexads-' + zoneId;
  container.style.cssText = 'display:inline-block;vertical-align:middle;';
  currentScript.parentNode.insertBefore(container, currentScript);

  // Fetch ad from adserve
  var adserveUrl = 'https://adserve.otexads.com/serve?zone=' + encodeURIComponent(zoneId);
  
  fetch(adserveUrl)
    .then(function(response) {
      if (response.status === 204) {
        // No ads available
        container.style.display = 'none';
        return null;
      }
      if (!response.ok) {
        throw new Error('Failed to fetch ad: ' + response.status);
      }
      return response.json();
    })
    .then(function(payload) {
      if (!payload) return;

      // The adserve API wraps responses in { success, data }. Unwrap if present.
      var ad = payload && payload.data ? payload.data : payload;
      if (!ad || !ad.format) return;

      // Render the ad based on format
      renderAd(container, ad);
    })
    .catch(function(error) {
      console.error('[OtexAds] Error loading ad:', error);
      container.style.display = 'none';
    });

  function renderAd(container, ad) {
    var clickUrl = 'https://adserve.otexads.com/click?token=' + encodeURIComponent(ad.click_token);
    var html = '';

    switch (ad.format) {
      case 'push':
        html = renderPushAd(ad, clickUrl);
        break;
      case 'native':
        html = renderNativeAd(ad, clickUrl);
        break;
      case 'banner':
        html = renderBannerAd(ad, clickUrl);
        break;
      case 'popunder':
        html = renderPopunderAd(ad, clickUrl);
        break;
      case 'interstitial':
        html = renderInterstitialAd(ad, clickUrl);
        break;
      default:
        html = renderBannerAd(ad, clickUrl);
    }

    container.innerHTML = html;

    // Track impression (already logged server-side, but we can add client-side validation if needed)
    // The impression is logged in adserve's /serve endpoint
  }

  function renderPushAd(ad, clickUrl) {
    var icon = ad.metadata && ad.metadata.push_icon ? ad.metadata.push_icon : ad.image_url;
    var title = ad.metadata && ad.metadata.push_title ? ad.metadata.push_title : ad.title;
    var body = ad.metadata && ad.metadata.push_body ? ad.metadata.push_body : ad.body;

    return '<a href="' + clickUrl + '" target="_blank" rel="noopener noreferrer" style="display:flex;align-items:center;text-decoration:none;color:inherit;font-family:-apple-system,BlinkMacSystemFont,Segoe UI,Roboto,sans-serif;max-width:400px;padding:12px;background:#f5f5f5;border-radius:8px;box-shadow:0 2px 4px rgba(0,0,0,0.1);">' +
      '<img src="' + icon + '" alt="" style="width:48px;height:48px;border-radius:4px;margin-right:12px;object-fit:cover;" />' +
      '<div style="flex:1;min-width:0;">' +
        '<div style="font-weight:600;font-size:14px;margin-bottom:4px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">' + escapeHtml(title) + '</div>' +
        '<div style="font-size:12px;color:#666;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">' + escapeHtml(body) + '</div>' +
      '</div>' +
    '</a>';
  }

  function renderNativeAd(ad, clickUrl) {
    var icon = ad.metadata && ad.metadata.icon_url ? ad.metadata.icon_url : ad.image_url;
    var cta = ad.metadata && ad.metadata.cta_text ? ad.metadata.cta_text : 'Learn More';
    var sponsored = ad.metadata && ad.metadata.sponsored_by ? ad.metadata.sponsored_by : 'Sponsored';

    return '<a href="' + clickUrl + '" target="_blank" rel="noopener noreferrer" style="display:flex;align-items:center;text-decoration:none;color:inherit;font-family:-apple-system,BlinkMacSystemFont,Segoe UI,Roboto,sans-serif;max-width:400px;padding:16px;background:#fff;border:1px solid #e0e0e0;border-radius:8px;">' +
      '<img src="' + icon + '" alt="" style="width:60px;height:60px;border-radius:4px;margin-right:12px;object-fit:cover;" />' +
      '<div style="flex:1;min-width:0;">' +
        '<div style="font-size:11px;color:#999;margin-bottom:4px;text-transform:uppercase;">' + escapeHtml(sponsored) + '</div>' +
        '<div style="font-weight:600;font-size:15px;margin-bottom:4px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">' + escapeHtml(ad.title) + '</div>' +
        '<div style="font-size:13px;color:#666;margin-bottom:8px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">' + escapeHtml(ad.body) + '</div>' +
        '<div style="display:inline-block;padding:6px 12px;background:#007bff;color:#fff;border-radius:4px;font-size:12px;font-weight:500;">' + escapeHtml(cta) + '</div>' +
      '</div>' +
    '</a>';
  }

  function renderBannerAd(ad, clickUrl) {
    var width = ad.metadata && ad.metadata.width ? ad.metadata.width : '300';
    var height = ad.metadata && ad.metadata.height ? ad.metadata.height : '250';

    return '<a href="' + clickUrl + '" target="_blank" rel="noopener noreferrer" style="display:inline-block;text-decoration:none;">' +
      '<img src="' + ad.image_url + '" alt="' + escapeHtml(ad.title) + '" style="width:' + width + 'px;height:' + height + 'px;border:0;" />' +
    '</a>';
  }

  function renderPopunderAd(ad, clickUrl) {
    // Popunder ads open in a new window/tab
    return '<a href="' + clickUrl + '" target="_blank" rel="noopener noreferrer" style="display:inline-block;text-decoration:none;">' +
      '<img src="' + ad.image_url + '" alt="' + escapeHtml(ad.title) + '" style="max-width:100%;height:auto;border:0;" />' +
    '</a>';
  }

  function renderInterstitialAd(ad, clickUrl) {
    var width = ad.metadata && ad.metadata.width ? ad.metadata.width : '800';
    var height = ad.metadata && ad.metadata.height ? ad.metadata.height : '600';

    return '<a href="' + clickUrl + '" target="_blank" rel="noopener noreferrer" style="display:inline-block;text-decoration:none;">' +
      '<img src="' + ad.image_url + '" alt="' + escapeHtml(ad.title) + '" style="width:' + width + 'px;height:' + height + 'px;border:0;" />' +
    '</a>';
  }

  function escapeHtml(text) {
    if (!text) return '';
    var div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }
})();
