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
        emitLogoutSuccessEvent,
        emitLogoutFailureEvent,
      } = useContext(AppContext);

    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [isSuccess, setIsSuccess] = useState<boolean>(false);

    const onLogout = (event: Event) => {
        event.preventDefault();
        setIsLoading(true);

        hanko.user.logout()
            .then(() => {
                setIsLoading(false);
                setIsSuccess(true);
                emitLogoutSuccessEvent();
            }).catch(() => {
                setIsLoading(false);
                emitLogoutFailureEvent();
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
