interface Auth {
  isLoggedIn: boolean;
}

export class Welcome {
  auth: Auth = {isLoggedIn: false};
}
