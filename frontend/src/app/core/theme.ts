import { Service, computed, effect, inject, signal } from '@angular/core';
import { DOCUMENT } from '@angular/common';

/** Theme preference: follow the OS (`system`) or force a scheme. */
export type ThemeMode = 'light' | 'dark' | 'system';

const STORAGE_KEY = 'theme-mode';

/**
 * Drives light/dark/system theming. Angular Material's M3 tokens (from
 * `mat.theme()`) resolve via the CSS `color-scheme` property, so switching the
 * theme is just a matter of setting `color-scheme` on the document element:
 * `light dark` follows the OS, `light`/`dark` force a scheme. The choice is
 * persisted to localStorage and reapplied on load.
 */
@Service()
export class Theme {
  private readonly document = inject(DOCUMENT);

  /** The active theme preference. */
  readonly mode = signal<ThemeMode>(this.readStoredMode());

  /** Material icon name reflecting the current mode (for the toolbar trigger). */
  readonly icon = computed(() => {
    switch (this.mode()) {
      case 'light':
        return 'light_mode';
      case 'dark':
        return 'dark_mode';
      default:
        return 'brightness_auto';
    }
  });

  constructor() {
    // Apply the scheme to <html> and persist the choice whenever it changes.
    effect(() => {
      const mode = this.mode();
      this.document.documentElement.style.colorScheme = mode === 'system' ? 'light dark' : mode;
      try {
        localStorage.setItem(STORAGE_KEY, mode);
      } catch {
        // Storage may be unavailable (private mode); the choice still applies this session.
      }
    });
  }

  /** Set the active theme mode. */
  set(mode: ThemeMode): void {
    this.mode.set(mode);
  }

  private readStoredMode(): ThemeMode {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored === 'light' || stored === 'dark' || stored === 'system') {
        return stored;
      }
    } catch {
      // Ignore storage errors and fall back to following the system.
    }
    return 'system';
  }
}
