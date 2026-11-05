import 'dart:async';
import 'dart:io';
import 'package:flutter/services.dart';

class OrcaException implements Exception {
  final String message;
  OrcaException(this.message);
  @override
  String toString() => message;
}

class OrcaBridge {
  static const _allowedCommands = {'init', 'add', 'list', 'get', 'gen'};
  static final _valuePattern = RegExp(r'^[A-Za-z0-9_.-]{1,256}$');

  static String? _binaryPath;

  static Future<String> _resolveBinary() async {
    if (_binaryPath != null) return _binaryPath!;
    if (Platform.isAndroid) {
      const channel = MethodChannel('orca/native');
      final dir = await channel.invokeMethod<String>('nativeLibraryDir');
      _binaryPath = '$dir/liborca.so';
    } else {
      _binaryPath = 'orca';
    }
    return _binaryPath!;
  }

  static void _validateArgs(List<String> args) {
    if (args.isEmpty) throw OrcaException('operation failed');
    if (!_allowedCommands.contains(args[0])) {
      throw OrcaException('operation failed');
    }
    for (final a in args.skip(1)) {
      if (!_valuePattern.hasMatch(a)) {
        throw OrcaException('operation failed');
      }
    }
  }

  static Future<String> run(List<String> args, {String? stdin}) async {
    _validateArgs(args);
    if (stdin != null && (stdin.contains('\n') || stdin.contains('\r'))) {
      // Passwords must not contain newlines (would break framing).
      throw OrcaException('operation failed');
    }
    final bin = await _resolveBinary();
    Process? proc;
    try {
      proc = await Process.start(bin, args).timeout(const Duration(seconds: 30));
      if (stdin != null) {
        proc.stdin.write(stdin);
        await proc.stdin.close();
      }
      final outFuture = proc.stdout.transform(const SystemEncoding().decoder).join();
      final errFuture = proc.stderr.transform(const SystemEncoding().decoder).join();
      final out = await outFuture.timeout(const Duration(seconds: 30));
      await errFuture.timeout(const Duration(seconds: 30));
      final code = await proc.exitCode.timeout(const Duration(seconds: 30));
      if (code != 0) {
        throw OrcaException('operation failed');
      }
      return out;
    } on TimeoutException {
      proc?.kill();
      throw OrcaException('operation timed out');
    } catch (_) {
      proc?.kill();
      throw OrcaException('operation failed');
    }
  }
}
