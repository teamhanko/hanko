// eslint-disable-next-line require-jsdoc
export function redirectTo(url: string) {
  window.location.assign(url);
}

// eslint-disable-next-line require-jsdoc
export function getCurrentHref(): string {
  return window.location.href;
}
