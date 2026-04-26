extends Control

const NEXT_SCENE := "res://scenes/main.tscn"
const SPLASH_DURATION := 5.0
const FADE_IN_DURATION := 1.2
const FADE_OUT_DURATION := 0.4
const BLINK_INTERVAL := 0.7

var _advanced := false
var _ready_for_input := false


func _ready() -> void:
	modulate.a = 0.0
	$Prompt.modulate.a = 0.0
	_fade_in()


func _fade_in() -> void:
	var tween := create_tween()
	tween.tween_property(self, "modulate:a", 1.0, FADE_IN_DURATION)
	tween.tween_callback(func() -> void:
		_ready_for_input = true
		_blink_prompt()
	)


func _blink_prompt() -> void:
	var tween := create_tween().set_loops()
	tween.tween_property($Prompt, "modulate:a", 1.0, BLINK_INTERVAL)
	tween.tween_property($Prompt, "modulate:a", 0.2, BLINK_INTERVAL)


func _input(event: InputEvent) -> void:
	if _ready_for_input and event.is_pressed():
		_advance()


func _advance() -> void:
	if _advanced:
		return
	_advanced = true
	set_process_input(false)
	var tween := create_tween()
	tween.tween_property(self, "modulate:a", 0.0, FADE_OUT_DURATION)
	tween.tween_callback(func() -> void:
		get_tree().change_scene_to_file(NEXT_SCENE)
	)
