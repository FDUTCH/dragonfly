package enchantment

import "github.com/df-mc/dragonfly/server/item"

func init() {
	item.RegisterEnchantment(IDProtection, Protection)
	item.RegisterEnchantment(IDFireProtection, FireProtection)
	item.RegisterEnchantment(IDFeatherFalling, FeatherFalling)
	item.RegisterEnchantment(IDBlastProtection, BlastProtection)
	item.RegisterEnchantment(IDProjectileProtection, ProjectileProtection)
	item.RegisterEnchantment(IDThorns, Thorns)
	item.RegisterEnchantment(IDRespiration, Respiration)
	item.RegisterEnchantment(IDDepthStrider, DepthStrider)
	item.RegisterEnchantment(IDAquaAffinity, AquaAffinity)
	item.RegisterEnchantment(IDSharpness, Sharpness)
	// TODO: (10) Smite. (Requires undead mobs)
	// TODO: (11) Bane of Arthropods. (Requires arthropod mobs)
	item.RegisterEnchantment(IDKnockback, Knockback)
	item.RegisterEnchantment(IDFireAspect, FireAspect)
	item.RegisterEnchantment(IDLooting, Looting)
	item.RegisterEnchantment(IDEfficiency, Efficiency)
	item.RegisterEnchantment(IDSilkTouch, SilkTouch)
	item.RegisterEnchantment(IDUnbreaking, Unbreaking)
	// TODO: (18) Fortune.
	item.RegisterEnchantment(IDPower, Power)
	item.RegisterEnchantment(IDPunch, Punch)
	item.RegisterEnchantment(IDFlame, Flame)
	item.RegisterEnchantment(IDInfinity, Infinity)
	// TODO: (23) Luck of the Sea.
	// TODO: (24) Lure.
	// TODO: (25) Frost Walker.
	item.RegisterEnchantment(IDMending, Mending)
	// TODO: (27) Curse of Binding.
	item.RegisterEnchantment(IDCurseOfVanishing, CurseOfVanishing)
	// TODO: (29) Impaling.
	// TODO: (30) Riptide.
	// TODO: (31) Loyalty.
	// TODO: (32) Channeling.
	// TODO: (33) Multishot.
	// TODO: (34) Piercing.
	item.RegisterEnchantment(IDQuickCharge, QuickCharge)
	item.RegisterEnchantment(IDSoulSpeed, SoulSpeed)
	item.RegisterEnchantment(IDSwiftSneak, SwiftSneak)
}
