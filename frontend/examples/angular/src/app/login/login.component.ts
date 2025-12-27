import { Component, CUSTOM_ELEMENTS_SCHEMA } from "@angular/core";
import { HankoService } from "../services/hanko.services";
import { Router } from "@angular/router";

@Component({
    selector: 'app-login',
    templateUrl: './login.component.html',
    styleUrls: ['../app.component.css'],
    standalone: true,
    schemas: [CUSTOM_ELEMENTS_SCHEMA]
})
export class LoginComponent {
  error?: Error;

  constructor(private hankoService: HankoService, private router: Router) {}

  redirectToTodos() {
    this.router.navigate(['/todo']).catch((e) => (this.error = e));
  }
}
