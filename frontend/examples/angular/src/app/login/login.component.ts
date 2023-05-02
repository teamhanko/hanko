import { Component, OnDestroy, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { HankoService } from "../services/hanko.services";

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['../app.component.css']
})
export class LoginComponent implements OnInit, OnDestroy {
  error: Error | undefined;

  constructor(private hankoService: HankoService, private router: Router) {}

  ngOnInit() {
    this.hankoService.register().catch((e) => this.error = e);
    this.hankoService.client.onAuthFlowCompleted(() => this.redirectToTodo())
  }

  ngOnDestroy() {
    this.hankoService.client.removeEventListeners();
  }

  redirectToTodo() {
    this.router.navigate(['/todo']).catch((e) => this.error = e);
  }
}
