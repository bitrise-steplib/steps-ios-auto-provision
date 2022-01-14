# iOS Auto Provision with Apple ID (Deprecated)

[![Step changelog](https://shields.io/github/v/release/bitrise-steplib/steps-ios-auto-provision?include_prereleases&label=changelog&color=blueviolet)](https://github.com/bitrise-steplib/steps-ios-auto-provision/releases)

Automatically manages your iOS Provisioning Profiles for your Xcode project.

<details>
<summary>Description</summary>

### This Step has been deprecated in favour of the new automatic code signing options on Bitrise.
You can read more about these changes in our blog post: [https://blog.bitrise.io/post/simplifying-automatic-code-signing-on-bitrise](https://blog.bitrise.io/post/simplifying-automatic-code-signing-on-bitrise).

#### Option A)
The latest versions of the [Xcode Archive & Export for iOS](https://www.bitrise.io/integrations/steps/xcode-archive), [Xcode Build for testing for iOS](https://www.bitrise.io/integrations/steps/xcode-build-for-test), and the [Export iOS and tvOS Xcode archive](https://www.bitrise.io/integrations/steps/xcode-archive) Steps have built-in automatic code signing.
We recommend removing this Step from your Workflow and using the automatic code signing feature in the Steps mentioned above.

#### Option B)
If you are not using any of the mentioned Xcode Steps, then you can replace
this iOS Auto Provision Step with the [Manage iOS Code signing](https://www.bitrise.io/integrations/steps/manage-ios-code-signing) Step.

### Description
The [Step](https://devcenter.bitrise.io/getting-started/configuring-bitrise-steps-that-require-apple-developer-account-data/#ios-auto-provision-with-apple-id-step) uses session-based authentication to connect to an Apple Developer account. In addition to an Apple ID and password, it also stores the 2-factor authentication (2FA) code you provide.

Please note that the [iOS Auto Provision with App Store Connect API](https://app.bitrise.io/integrations/steps/ios-auto-provision-appstoreconnect) Step uses the official [App Store Connect API](https://developer.apple.com/documentation/appstoreconnectapi/generating_tokens_for_api_requests) instead of the old session-based method.

The **iOS Auto Provision with Apple ID** Step supports in Xcode managed and manual code signing in the following ways:

In the case of Xcode managed code signing projects, the Step:
- Downloads the Xcode managed Provisioning Profiles and installs them for the build.
- Installs the provided code signing certificates into the Keychain.
In the case of manual code signing projects, the Step:
- Ensures that the Application Identifier exists on the Apple Developer Portal.
- Ensures that the project's Capabilities are set correctly in the Application Identifier.
- Ensures that the Provisioning Profiles exist on the Apple Developer Portal and are installed for the build.
- Ensures that all the available Test Devices exist on the Apple Developer Portal and are included in the Provisioning Profiles.
- Installs the provided code signing certificates into the Keychain.

### Configuring the Step

Before you start configuring the Step, make sure you've completed the following requirements:
- You've [defining your Apple Developer Account to Bitrise](https://devcenter.bitrise.io/getting-started/configuring-bitrise-steps-that-require-apple-developer-account-data/#defining-your-apple-developer-account-to-bitrise-1).
- You've [assigned an Apple Developer Account for your app](https://devcenter.bitrise.io/getting-started/configuring-bitrise-steps-that-require-apple-developer-account-data/#assigning-an-apple-developer-account-for-your-app-1).

To configure the Step:
Once you've completed the above requirements, there is very little configuration needed to this Step.
1. Add the **iOS Auto Provision with Apple ID** Step after any dependency installer Step in your Workflow, such as **Run CocoaPods install** or **Carthage**.
2. Click the Step to edit its input fields. You can see that the **Distribution type**, **Xcode Project (or Workspace) path**, and the **Scheme name** inputs are automatically filled out for you.
3. If your Developer Portal Account belongs to multiple development teams, add the **Developer Portal team ID** to manage the project's code signing files, for example '1MZX23ABCD4'. If that's not the case, you can still add it to manage the Provisioning Profiles with a different team than the one set in your project. If you leave it empty, the team defined by the project will be used.
4. If you wish to overwrite the configuration defined in your Scheme (for example, Debug, Release), you can do so in the **Configuration name** input.
5. If Xcode managed signing is enabled in the iOS app, check the value of the **Should the step try to generate Provisioning Profiles even if Xcode managed signing is enabled in the Xcode project?** input.
- If it’s set to 'no', the Step will look for an Xcode Managed Provisioning Profile on the Apple Developer Portal.
- If it’s set to 'yes', the Step will generate a new manual provisioning profile on the Apple Developer portal for the project.
This input has no effect in the case of Manual code signing projects.
6. **The minimum days the Provisioning Profile should be valid** lets you specify how long a Provisioning Profile should be valid to sign an iOS app. By default it will only renew the Provisioning Profile when it expires.

### Troubleshooting
Please note that the 2FA code is only valid for 30 days. 
When the 2FA code expires, you will need to re-authenticate to provide a new code. 
Go to the Apple Developer Account of the **Account settings** page, it will automatically ask for the 2FA code to authenticate again. 
There will be a list of the Apple Developer accounts that you have defined. To the far right of each, there are 3 dots. 
Click the dots and select **Re-authenticate (2SA/2FA)**.

### Useful links
- [Managing code signing files - automatic provisioning](https://devcenter.bitrise.io/code-signing/ios-code-signing/ios-auto-provisioning/#configuring-ios-auto-provisioning)
- [iOS code signing troubleshooting](https://devcenter.bitrise.io/code-signing/ios-code-signing/ios-code-signing-troubleshooting/)

### Related Steps
- [iOS Auto Provision with App Store Connect API](https://app.bitrise.io/integrations/steps/ios-auto-provision-appstoreconnect)
- [Xcode Archive & Export](https://www.bitrise.io/integrations/steps/xcode-archive)
</details>

## 🧩 Get started

Add this step directly to your workflow in the [Bitrise Workflow Editor](https://devcenter.bitrise.io/steps-and-workflows/steps-and-workflows-index/).

You can also run this step directly with [Bitrise CLI](https://github.com/bitrise-io/bitrise).

## ⚙️ Configuration

<details>
<summary>Inputs</summary>

| Key | Description | Flags | Default |
| --- | --- | --- | --- |
| `distribution_type` | Describes how Xcode should sign your project. | required | `development` |
| `team_id` | The Developer Portal team to manage the project's code signing files. __If your Developer Portal Account belongs to multiple development team, this input is required!__ Otherwise specify this input if you want to manage the Provisioning Profiles with a different team than the one set in your project. If you leave it empty the team defined by the project will be used. __Example:__ `1MZX23ABCD4` |  |  |
| `project_path` | A `.xcodeproj` or `.xcworkspace` path. | required | `$BITRISE_PROJECT_PATH` |
| `scheme` | The Xcode Scheme to use. | required | `$BITRISE_SCHEME` |
| `configuration` | The Xcode Configuration to use. By default your Scheme defines which Configuration (Debug, Release, ...) should be used, but you can overwrite it with this option. |  |  |
| `generate_profiles` | In the case of __Xcode managed code signing__ projects, by default the step downloads and installs the Xcode managed Provisioning Profiles. If this input is set to: `yes`, the step will try to manage the Provisioning Profiles by itself (__like in the case of Manual code signing projects__), the step will fall back to use the Xcode managed Provisioning Profiles if there is an issue. __This input has no effect in the case of Manual codesigning projects.__ |  | `no` |
| `register_test_devices` | If set the step will register known test devices on Bitrise from team members with the Apple Developer Portal. Note that setting this to "yes" may cause devices to be registered against your limited quantity of test devices in the Apple Developer Portal, which can only be removed once annually during your renewal window. |  | `no` |
| `min_profile_days_valid` | Sometimes you want to sign an app with a Provisioning Profile that is valid for at least 'x' days. For example, an enterprise app won't open if your Provisioning Profile is expired. With this parameter, you can have a Provisioning Profile that's at least valid for 'x' days.  By default (0) it just renews the Provisioning Profile when expired. |  | `0` |
| `verbose_log` | Enable verbose logging? | required | `no` |
| `certificate_urls` | URLs of the certificates to download. Multiple URLs can be specified, separated by a pipe (`\|`) character, you can specify a local path as well, using the `file://` scheme. __Provide a development certificate__ url, to ensure development code signing files for the project and __also provide a distribution certificate__ url, to ensure distribution code signing files for your project. __Example:__ `file://./development/certificate/path\|https://distribution/certificate/url`  | required, sensitive | `$BITRISE_CERTIFICATE_URL` |
| `passphrases` | Certificate passphrases. Multiple passphrases can be specified, separated by a pipe (`\|`) character. __Specified certificate passphrase count should match the count of the certificate URLs.__ For example, (1 certificate with empty passphrase, 1 certificate with non-empty passphrase) `\|distribution-passphrase`.  | required, sensitive | `$BITRISE_CERTIFICATE_PASSPHRASE` |
| `keychain_path` | The Keychain path. | required | `$HOME/Library/Keychains/login.keychain` |
| `keychain_password` | The Keychain's password. | required, sensitive | `$BITRISE_KEYCHAIN_PASSWORD` |
| `build_url` | Bitrise build URL. | required | `$BITRISE_BUILD_URL` |
| `build_api_token` | Bitrise build API token. | required, sensitive | `$BITRISE_BUILD_API_TOKEN` |
</details>

<details>
<summary>Outputs</summary>

| Environment Variable | Description |
| --- | --- |
| `BITRISE_EXPORT_METHOD` | The selected distribution type. One of these: `development`, `app-store`, `ad-hoc` or `enterprise`. |
| `BITRISE_DEVELOPER_TEAM` | The development team's ID. Example: `1MZX23ABCD4` |
| `BITRISE_DEVELOPMENT_CODESIGN_IDENTITY` | The development code signing identity's name. For example, `iPhone Developer: Bitrise Bot (VV2J4SV8V4)`. |
| `BITRISE_PRODUCTION_CODESIGN_IDENTITY` | The production code signing identity's name. Example: `iPhone Distribution: Bitrise Bot (VV2J4SV8V4)` |
| `BITRISE_DEVELOPMENT_PROFILE` | The main target's development provisioning profile's UUID. Example: `c5be4123-1234-4f9d-9843-0d9be985a068` |
| `BITRISE_PRODUCTION_PROFILE` | The main target's production provisioning profile UUID. Example: `c5be4123-1234-4f9d-9843-0d9be985a068` |
</details>

## 🙋 Contributing

We welcome [pull requests](https://github.com/bitrise-steplib/steps-ios-auto-provision/pulls) and [issues](https://github.com/bitrise-steplib/steps-ios-auto-provision/issues) against this repository.

For pull requests, work on your changes in a forked repository and use the Bitrise CLI to [run step tests locally](https://devcenter.bitrise.io/bitrise-cli/run-your-first-build/).

**Note:** this step's end-to-end tests (defined in `e2e/bitrise.yml`) are working with secrets which are intentionally not stored in this repo. External contributors won't be able to run those tests. Don't worry, if you open a PR with your contribution, we will help with running tests and make sure that they pass.

Learn more about developing steps:

- [Create your own step](https://devcenter.bitrise.io/contributors/create-your-own-step/)
- [Testing your Step](https://devcenter.bitrise.io/contributors/testing-and-versioning-your-steps/)
