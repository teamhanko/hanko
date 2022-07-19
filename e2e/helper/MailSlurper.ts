import fetch from "node-fetch";

export interface Mails {
  mailItems: Mail[];
}

export interface Mail {
  id: string;
  dateSent: string;
  fromAddress: string;
  toAddresses: string[];
  subject: string;
  xmailer: string;
  body: string;
  contentType: string;
  boundary: string;
  attachments: Attachment[];
}

export interface Attachment {
  id: string;
  mailId: string;
  headers: Headers;
  contents: string;
}

export interface Headers {
  contentType: string;
  mimeVersion: string;
  contentTransferEncoding: string;
  contentDisposition: string;
  fileName: string;
  body: string;
}

export class MailSlurper {
  api: string;

  constructor(protocol = "http", host = "localhost", port = 8085) {
    this.api = `${protocol}://${host}:${port}`;
  }

  async getMails(recipient: string): Promise<Mails> {
    const url = recipient
      ? `${this.api}/mail?to=${recipient}`
      : `${this.api}/mail`;
    const response = await fetch(url);
    return (await response.json()) as Mails;
  }

  async getPasscodeFromMail(mail: Mail): Promise<string> {
    const passcode = mail.body.match("[0-9]{6}");

    if (passcode) {
      return passcode[0];
    } else {
      throw new Error(
        `could not extract passcode from mail with id'${mail.id}'`
      );
    }
  }

  async getPasscodeFromMostRecentMail(recipient: string): Promise<string> {
    const mails: Mails = await this.getMails(recipient);

    if (mails.mailItems.length === 0) {
      throw new Error(`no mails found for '${recipient}'`);
    }

    return this.getPasscodeFromMail(mails.mailItems[0]);
  }
}
