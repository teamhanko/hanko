import * as preact from "preact";
import { useContext, useState } from "preact/compat";

import { AppContext } from "../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Button from "../components/form/Button";
type Props = {
};

const LogoutButton = (props: Props) => {
    const { t } = useContext(TranslateContext);
    const {
        hanko,
        emitLogoutEvent,
      } = useContext(AppContext);

    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [isSuccess, setIsSuccess] = useState<boolean>(false);

    const onLogout = (event: Event) => {
        event.preventDefault();
        setIsLoading(true);

        hanko.user.logout()
            .then((resp: boolean) => {
                setIsLoading(false);

                if (resp) {
                    // TODO: This is not working. How can I navigate back to the home screen?
                    // setPage(<LoginEmailPage />);
                    setIsSuccess(true);
                    emitLogoutEvent();
                    return;
                }
            });
    }

    return (
        <Content>
            <Form onSubmit={onLogout}>
                <Button isLoading={isLoading} isSuccess={isSuccess}>
                    {t("labels.logout")}
                </Button>
            </Form>
        </Content>
    )
}

export default LogoutButton;
