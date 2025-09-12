import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';

import { Button } from '@@/buttons';
import { FormControl } from '@@/form-components/FormControl';
import { Input } from '@@/form-components/Input';
import { Widget, WidgetBody, WidgetTitle } from '@@/Widget';
import { notifySuccess, notifyError } from '@@/notifications';
import { withInvalidate, queryClient } from '@/react-tools/react-query';

import { Settings, PGPSettings } from '../types';
import { updateSettings } from '../settings.service';

interface Props {
  settings: Settings;
}

export function PGPSettingsPanel({ settings }: Props) {
  const [pgpSettings, setPGPSettings] = useState<PGPSettings>(
    settings.PGPSettings || { PrivateKey: '', Passphrase: '' }
  );
  const [isPasswordVisible, setIsPasswordVisible] = useState(false);

  const updateSettingsMutation = useMutation(updateSettings, {
    onSuccess() {
      notifySuccess('Success', 'PGP settings saved successfully');
      queryClient.invalidateQueries(['settings']);
    },
    onError(error: Error) {
      notifyError('Failure', error, 'Failed to save PGP settings');
    },
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      await updateSettingsMutation.mutateAsync({
        PGPSettings: pgpSettings,
      });
    } catch (error) {
      // Error is handled by the mutation
    }
  };

  const handlePrivateKeyChange = (value: string) => {
    setPGPSettings((prev) => ({
      ...prev,
      PrivateKey: value,
    }));
  };

  const handlePassphraseChange = (value: string) => {
    setPGPSettings((prev) => ({
      ...prev,
      Passphrase: value,
    }));
  };

  const clearSettings = () => {
    setPGPSettings({ PrivateKey: '', Passphrase: '' });
  };

  return (
    <Widget>
      <WidgetTitle icon="lock" title="PGP Settings" />
      <WidgetBody>
        <div className="form-horizontal">
          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <div className="col-sm-12">
                <div className="alert alert-info">
                  <i className="fa fa-info-circle" aria-hidden="true" />
                  Configure PGP settings for automatically decrypting encrypted secrets in GitOps stacks.
                  These settings are used when the <code>PORTAINER_PGP_PRIVATE_KEY</code> environment variable is not set.
                </div>
              </div>
            </div>

            <FormControl
              label="PGP Private Key"
              tooltip="ASCII armored PGP private key for decrypting secrets"
            >
              <textarea
                className="form-control"
                rows={10}
                value={pgpSettings.PrivateKey || ''}
                onChange={(e) => handlePrivateKeyChange(e.target.value)}
                placeholder="-----BEGIN PGP PRIVATE KEY BLOCK-----&#10;...&#10;-----END PGP PRIVATE KEY BLOCK-----"
                data-cy="pgp-private-key"
              />
            </FormControl>

            <FormControl
              label="Passphrase"
              tooltip="Optional passphrase for the private key"
            >
              <div className="input-group">
                <Input
                  type={isPasswordVisible ? 'text' : 'password'}
                  value={pgpSettings.Passphrase || ''}
                  onChange={(e) => handlePassphraseChange(e.target.value)}
                  placeholder="Enter passphrase (leave empty if key has no passphrase)"
                  data-cy="pgp-passphrase"
                />
                <div className="input-group-append">
                  <Button
                    size="small"
                    color="secondary"
                    type="button"
                    onClick={() => setIsPasswordVisible(!isPasswordVisible)}
                    data-cy="toggle-passphrase-visibility"
                  >
                    <i
                      className={`fa ${
                        isPasswordVisible ? 'fa-eye-slash' : 'fa-eye'
                      }`}
                      aria-hidden="true"
                    />
                  </Button>
                </div>
              </div>
            </FormControl>

            <div className="form-group">
              <div className="col-sm-12">
                <div className="btn-toolbar" role="toolbar">
                  <div className="btn-group" role="group">
                    <Button
                      type="submit"
                      disabled={updateSettingsMutation.isLoading}
                      data-cy="save-pgp-settings"
                    >
                      {updateSettingsMutation.isLoading ? (
                        <>
                          <i className="fa fa-spinner fa-spin" /> Saving...
                        </>
                      ) : (
                        'Save settings'
                      )}
                    </Button>
                  </div>
                  <div className="btn-group" role="group">
                    <Button
                      type="button"
                      color="danger"
                      onClick={clearSettings}
                      disabled={updateSettingsMutation.isLoading}
                      data-cy="clear-pgp-settings"
                    >
                      Clear settings
                    </Button>
                  </div>
                </div>
              </div>
            </div>

            <div className="form-group">
              <div className="col-sm-12">
                <div className="alert alert-warning">
                  <i className="fa fa-exclamation-triangle" aria-hidden="true" />
                  <strong>Security Notice:</strong> Private keys and passphrases are stored encrypted in the database.
                  Environment variable <code>PORTAINER_PGP_PRIVATE_KEY</code> takes precedence over these settings.
                </div>
              </div>
            </div>
          </form>
        </div>
      </WidgetBody>
    </Widget>
  );
}