include <BOSL2/std.scad>;

text_str = "GREG";
font = "Baskerville:style=Bold";
font_size = 40;
text_height = 5;

// LED platform and groove parameters
led_block_length = 100;
led_block_width = 20;
led_block_height = 6;

led_groove_width = 10;
led_groove_depth = 2;

module led_flat_groove_block(length, block_width, block_height, groove_width, groove_depth) {
    attachable(anchor = TOP) {
        difference() {
            // Solid LED block
            cuboid([length, block_width, block_height], anchor = BOTTOM);
            // Shallow groove on top for LED strip to sit in
          //  translate([0, 0, block_height - groove_depth])
          
                #cuboid([length+2,  groove_depth*2, groove_width], anchor = FWD)
                show_anchors(flag=true);
        }
        children();
    }
}

module text_block(str, font, size, height) {
    attachable(anchor = BOTTOM) {
        text3d(str,
            size = size,
            height = height,
            font = font,
            halign = "center",
            valign = "baseline",
            spacing = 0.9
        );
        children();
    }
}

// Combine: LED base with shallow top groove + text directly on top
led_flat_groove_block(led_block_length, led_block_height,led_block_width, led_groove_width, led_groove_depth)
attach(TOP)
    text_block(text_str, font, font_size, text_height);
