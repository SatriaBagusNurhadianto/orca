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
  bool _unlocked = false;

  Future<void> _unlock() async {
    try {
      final pw = _ctrl.text;
      final out = await OrcaBridge.run(['list'], stdinFields: [pw]);
      _ctrl.clear();
      setState(() {
        _entries = out.trim().split('\n').where((e) => e.isNotEmpty).toList();
        _status = 'Unlocked';
        _unlocked = true;
      });
    } catch (e) {
      _ctrl.clear();
      setState(() => _status = 'Error: $e');
    }
  }

  Future<void> _addEntry() async {
    final result = await Navigator.push<Map<String, String>>(
      context,
      MaterialPageRoute(builder: (_) => const AddEntryScreen()),
    );
    if (result == null) return;
    // Re-prompt master password via a simple dialog.
    final pw = await _promptPassword();
    if (pw == null) return;
    try {
      await OrcaBridge.run(['add'], stdinFields: [pw, result['title']!, result['username']!, result['password']!, result['url']!, result['notes']!]);
      final out = await OrcaBridge.run(['list'], stdinFields: [pw]);
      setState(() {
        _entries = out.trim().split('\n').where((e) => e.isNotEmpty).toList();
        _status = 'Entry added';
      });
    } catch (e) {
      setState(() => _status = 'Error: $e');
    }
  }

  Future<String?> _promptPassword() async {
    final ctrl = TextEditingController();
    return showDialog<String>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Master password'),
        content: TextField(controller: ctrl, obscureText: true),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel')),
          TextButton(onPressed: () => Navigator.pop(ctx, ctrl.text), child: const Text('OK')),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Orca Vault')),
      floatingActionButton: _unlocked
          ? FloatingActionButton(onPressed: _addEntry, child: const Icon(Icons.add))
          : null,
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            if (!_unlocked) ...[
              TextField(
                controller: _ctrl,
                obscureText: true,
                decoration: const InputDecoration(labelText: 'Master password'),
              ),
              const SizedBox(height: 12),
              ElevatedButton(onPressed: _unlock, child: const Text('Unlock')),
            ],
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

class AddEntryScreen extends StatefulWidget {
  const AddEntryScreen({super.key});
  @override
  State<AddEntryScreen> createState() => _AddEntryScreenState();
}

class _AddEntryScreenState extends State<AddEntryScreen> {
  final _title = TextEditingController();
  final _user = TextEditingController();
  final _pass = TextEditingController();
  final _url = TextEditingController();
  final _notes = TextEditingController();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Add Entry')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            TextField(controller: _title, decoration: const InputDecoration(labelText: 'Title')),
            TextField(controller: _user, decoration: const InputDecoration(labelText: 'Username')),
            TextField(controller: _pass, decoration: const InputDecoration(labelText: 'Password'), obscureText: true),
            TextField(controller: _url, decoration: const InputDecoration(labelText: 'URL')),
            TextField(controller: _notes, decoration: const InputDecoration(labelText: 'Notes')),
            const SizedBox(height: 12),
            ElevatedButton(
              onPressed: () => Navigator.pop(context, {
                'title': _title.text,
                'username': _user.text,
                'password': _pass.text,
                'url': _url.text,
                'notes': _notes.text,
              }),
              child: const Text('Save'),
            ),
          ],
        ),
      ),
    );
  }
}
