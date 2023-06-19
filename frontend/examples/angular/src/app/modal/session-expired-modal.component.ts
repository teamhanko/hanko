import { Component, ElementRef, ViewChild } from "@angular/core";
import { Router } from '@angular/router';
import { HankoService } from "../services/hanko.services";

@Component({
  selector: 'app-session-expired-modal',
  templateUrl: './session-expired-modal.component.html',
  styleUrls: ['../app.component.css'],
})
export class SessionExpiredModalComponent {
  @ViewChild('modal') modal?: ElementRef<HTMLDialogElement>;
  error?: Error;

  constructor(private hankoService: HankoService, private router: Router) {}

  redirectToLogin() {
    this.router.navigate(['/']).catch((e) => (this.error = e));
  }

  show() {
    this.modal?.nativeElement.showModal();
  }
}
