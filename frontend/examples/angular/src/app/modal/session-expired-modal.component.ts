import { Component, OnDestroy, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { HankoService } from "../services/hanko.services";

@Component({
  selector: 'app-modal',
  templateUrl: './modal.component.html',
  styleUrls: ['../app.component.css'],
})
export class ModalComponent implements OnInit, OnDestroy {
  error: Error | undefined;

  constructor(private hankoService: HankoService, private router: Router) {}

  ngOnInit() {
    this.hankoService.client.onSessionExpired(() => {
      this.openSessionExpiredModal();
    });
  }

  ngOnDestroy() {
    this.hankoService.client.removeEventListeners();
  }

  redirectToLogin() {
    this.router.navigate(['/']).catch((e) => (this.error = e));
  }

  openSessionExpiredModal() {
    const dialog = <HTMLDialogElement>document.querySelector('#session-expired-modal');
    dialog.showModal();
  }
}
