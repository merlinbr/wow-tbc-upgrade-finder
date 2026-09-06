// ZAM viewer adapter: provider-specific creation, updates, controls and
// disposal. Hides provider globals (window.ZamModelViewer, jQuery, WH, the
// CONTENT_PATH convention) and resource ownership inside this module.
import './vendor/jquery-1.12.4.min.js';
import { appearanceOptions } from './characterPreview.js';

export const ZAM_CONTENT_PATH = '/visuals/zam/modelviewer/tbc/';
const ZAM_VIEWER_URL = 'https://wow.zamimg.com/modelviewer/live/viewer/viewer.min.js';

let depsPromise = null;

function loadScript(src) {
  return new Promise((resolve, reject) => {
    const existing = document.querySelector(`script[src="${src}"]`);
    if (existing) {
      if (window.ZamModelViewer) return resolve();
      existing.addEventListener('load', resolve, { once: true });
      existing.addEventListener('error', () => reject(new Error('viewer script failed to load')), { once: true });
      return;
    }
    const script = document.createElement('script');
    script.src = src;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error('viewer script failed to load'));
    document.head.appendChild(script);
  });
}

// The viewer expects the reference site's helper globals.
function ensureShims() {
  window.WH = window.WH || {};
  window.WH.debug = window.WH.debug || function () {};
  window.WH.WebP = window.WH.WebP || { getImageExtension: () => '.webp' };
}

// Loads jQuery + the ZAM viewer script exactly once per session.
function loadViewerDependencies() {
  if (!depsPromise) {
    depsPromise = Promise.all([
      typeof window.jQuery === 'function' ? Promise.resolve() : import('./vendor/jquery-1.12.4.min.js'),
      loadScript(ZAM_VIEWER_URL),
    ])
      .then(() => {
        if (typeof window.ZamModelViewer !== 'function') {
          throw new Error('ZamModelViewer constructor not available');
        }
        ensureShims();
        return window.ZamModelViewer;
      })
      .catch((error) => {
        depsPromise = null; // allow retry after transient network failure
        throw error;
      });
  }
  return depsPromise;
}

export const zamAdapter = {
  kind: 'zam',

  /**
   * Creates a viewer for the given character.
   * @param {object} config
   * @param {HTMLElement} config.container  mount element
   * @param {number} config.raceId          ZAM race id
   * @param {0|1} config.gender             ZAM gender (0 male, 1 female)
   * @param {Array<[number, number]>} config.items  [zamSlot, displayId] pairs
   * @param {object} config.appearance      appearance indexes (defaults apply)
   * @returns {Promise<object>} viewer handle: rotate, resetView, resize, destroy
   */
  async create(config) {
    const ZamModelViewer = await loadViewerDependencies();
    const { container, raceId, gender, items, appearance } = config;
    const modelId = raceId * 2 - 1 + (gender === 1 ? 1 : 0);

    window.CONTENT_PATH = ZAM_CONTENT_PATH;
    const customizationResponse = await fetch(`${ZAM_CONTENT_PATH}meta/charactercustomization/${modelId}.json`);
    if (!customizationResponse.ok) {
      throw new Error(`character customization metadata unavailable (HTTP ${customizationResponse.status})`);
    }
    const customization = await customizationResponse.json();
    const options = appearanceOptions(appearance, customization);

    // The constructor returns a thenable that must be awaited.
    const viewer = await new ZamModelViewer({
      type: 2,
      contentPath: ZAM_CONTENT_PATH,
      container: window.jQuery(container),
      aspect: 0.79,
      sheatheWeapons: 0,
      autoSheathe: 0,
      hd: true,
      items,
      models: { id: modelId, type: 16 },
      charCustomization: { options },
    });
    fitMargin(viewer);

    const initialAzimuth = viewer.renderer.azimuth;
    const initialZenith = viewer.renderer.zenith;

    return {
      rotate(radians) {
        viewer.renderer.azimuth += radians;
      },
      resetView() {
        // 0/0 is not the default view on this viewer revision; restoring the
        // captured initial angles keeps the character visible after reset.
        viewer.renderer.azimuth = initialAzimuth;
        viewer.renderer.zenith = initialZenith;
      },
      resize(width, height) {
        viewer.renderer.resize(width, height);
      },
      destroy() {
        viewer.destroy();
      },
    };
  },
};

// The viewer auto-fits the model to the canvas once its assets finish
// loading; add a breathing margin so feet/weapon are never clipped.
// Measured in the real app at the 478x670 stage: distance * 1.4 frames the
// full body with top/bottom margins. Applied after the fit settles (the
// actor appears, then a short wait).
function fitMargin(viewer) {
  const started = Date.now();
  const timer = setInterval(() => {
    const renderer = viewer.renderer;
    if (!renderer) {
      if (Date.now() - started > 25000) clearInterval(timer);
      return;
    }
    const actors = renderer.actors;
    const ready = actors && actors.length > 0;
    if (ready || Date.now() - started > 25000) {
      clearInterval(timer);
      const apply = () => {
        if (renderer.actors?.length && renderer.distance) {
          renderer.distance *= 1.4;
        }
      };
      if (ready) setTimeout(apply, 2000);
    }
  }, 100);
}
