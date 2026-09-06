// Test-only fake viewer adapter: deterministic, no network. Mirrors the ZAM
// adapter's surface so CI can exercise the full stage activation flow
// (loading, ready, controls, cleanup) without provider access.
export const fakeAdapter = {
  kind: 'fake',

  async create(config) {
    const { container, raceId, gender } = config;
    const panel = document.createElement('div');
    panel.className = 'fake-viewer';
    panel.setAttribute('data-testid', 'fake-viewer');
    panel.textContent = `Fake 3D preview — race ${raceId}, ${gender === 1 ? 'female' : 'male'}`;
    container.appendChild(panel);
    let width = container.clientWidth;
    let height = container.clientHeight;
    return {
      rotate() {
        panel.style.transform = `rotate(${(Math.random() * 10).toFixed(1)}deg)`;
      },
      resetView() {
        panel.style.transform = 'none';
      },
      resize(w, h) {
        width = w;
        height = h;
        panel.style.minWidth = `${w}px`;
        panel.style.minHeight = `${h}px`;
      },
      destroy() {
        panel.remove();
      },
    };
  },
};
