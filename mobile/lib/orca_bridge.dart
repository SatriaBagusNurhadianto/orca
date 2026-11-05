import 'dart:io';
import 'package:flutter/services.dart';

class OrcaBridge {
  static String? _binaryPath;

  static Future<String> _resolveBinary() async {
    if (_binaryPath != null) return _binaryPath!;
    // On Android, native libs are extracted to nativeLibraryDir.
    if (Platform.isAndroid) {
      // Use path_provider? nativeLibraryDir not directly available in Dart.
      // Use a platform channel to get nativeLibraryDir.
      const channel = MethodChannel('orca/native');
      final dir = await channel.invokeMethod<String>('nativeLibraryDir');
      _binaryPath = '$dir/liborca.so';
    } else {
      _binaryPath = 'orca';
    }
    return _binaryPath!;
  }

  static Future<String> run(List<String> args, {String? stdin}) async {
    final bin = await _resolveBinary();
    final proc = await Process.start(bin, args);
    if (stdin != null) {
      proc.stdin.write(stdin);
      await proc.stdin.close();
    }
    final out = await proc.stdout.transform(const SystemEncoding().decoder).join();
    final err = await proc.stderr.transform(const SystemEncoding().decoder).join();
    final code = await proc.exitCode;
    if (code != 0) {
      throw Exception('orca failed ($code): $err');
    }
    return out;
  }
}
