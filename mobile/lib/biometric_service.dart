import 'package:local_auth/local_auth.dart';

class BiometricService {
  final _auth = LocalAuthentication();

  Future<bool> canCheck() async {
    try {
      return await _auth.canCheckBiometrics && await _auth.isDeviceSupported();
    } catch (_) {
      return false;
    }
  }

  /// Returns true only on successful strong-biometric auth.
  /// Any failure/cancellation/unavailable returns false (fail closed).
  Future<bool> authenticate() async {
    if (!await canCheck()) return false;
    try {
      return await _auth.authenticate(
        localizedReason: 'Unlock Orca vault',
        options: const AuthenticationOptions(
          biometricOnly: true,
          stickyAuth: true,
        ),
      );
    } on Exception {
      return false;
    }
  }
}
