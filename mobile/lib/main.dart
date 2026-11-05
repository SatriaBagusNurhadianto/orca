import 'package:flutter/material.dart';
import 'orca_bridge.dart';

void main() => runApp(const OrcaApp());

class OrcaApp extends StatelessWidget {
  const OrcaApp({super.key});
  @override
  Widget build(BuildContext context) => MaterialApp(
        title: 'Orca',
        home: const UnlockScreen(),
      );
}

class UnlockScreen extends StatefulWidget {
  const UnlockScreen({super.key});
  @override
  State<UnlockScreen> createState() => _UnlockScreenState();
}

class _UnlockScreenState extends State<UnlockScreen> {
  final _ctrl = TextEditingController();
  String _status = '';
  List<String> _entries = [];

  Future<void> _unlock() async {
    try {
      final out = await OrcaBridge.run(['list'], stdin: '${_ctrl.text}\n');
      setState(() {
        _entries = out.trim().split('\n').where((e) => e.isNotEmpty).toList();
        _status = 'Unlocked';
      });
    } catch (e) {
      setState(() => _status = 'Error: $e');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Orca Vault')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            TextField(
              controller: _ctrl,
              obscureText: true,
              decoration: const InputDecoration(labelText: 'Master password'),
            ),
            const SizedBox(height: 12),
            ElevatedButton(onPressed: _unlock, child: const Text('Unlock')),
            const SizedBox(height: 12),
            Text(_status),
            Expanded(
              child: ListView(
                children: _entries.map((e) => ListTile(title: Text(e))).toList(),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
