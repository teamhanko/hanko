import { Injectable } from "@angular/core";
import { environment } from "../../environments/environment";
import { Hanko, register } from "@teamhanko/hanko-elements";

@Injectable({
  providedIn: 'root',
})
export class HankoService {
  api = environment.hankoApi
  client = new Hanko(this.api)

  constructor() {
    register(this.api).catch(console.error);
  }
}
