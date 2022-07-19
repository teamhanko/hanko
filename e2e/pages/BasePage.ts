import type { Locator, Page } from "@playwright/test";

export abstract class BasePage {
  readonly page: Page;
  readonly errorMessage: Locator;

  constructor(page: Page) {
    this.page = page;
    this.errorMessage = page.locator("id=errorMessage");
  }

  async hasCookie(name = "hanko") {
    const cookies = await this.page.context().cookies();
    const found = cookies.find((cookie) => cookie.name === name);
    return !!found;
  }

  async getLocalStorageValue(origin = "http://localhost:8888", key = "hanko") {
    const storageState = await this.page.context().storageState();
    const store = storageState.origins.find((o) => o.origin === origin);
    const entry = store?.localStorage.find((entry) => entry.name == key);
    if (entry) {
      return entry.value;
    } else {
      return null;
    }
  }

  async getDecodedLocalStorageValue(
    origin = "http://localhost:8888",
    key = "hanko"
  ) {
    const value = await this.getLocalStorageValue(origin, key);
    if (value) {
      const buf = new Buffer(value, "base64").toString();
      return JSON.parse(decodeURIComponent(decodeURI(buf)));
    } else {
      return null;
    }
  }
}
