import { Injectable } from "@angular/core";
import { environment } from "../../environments/environment";
import { createHankoClient, Hanko, register } from "@teamhanko/hanko-elements";

@Injectable({
  providedIn: 'root',
})
export class HankoService {
  api = environment.hankoApi
  client: Hanko = createHankoClient(this.api)
  register = () => register(this.api);
}
