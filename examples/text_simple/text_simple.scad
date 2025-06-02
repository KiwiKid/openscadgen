include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;


content="#BizSlutCon2025";
text_size=40;
h=5;
font="Didot:style=Italic";
spacing=1.0;
center=false;
spin=0;
orient=[0,0,-1];
hasBase = false;

module text_simple(
    content,
    text_size,
    h,
    font,
    spacing,
    center,
    spin,
    orient,
    hasBase,
baseHeight
){
    if(hasBase){
    difference(){
    
        #cuboid([500,text_size*1.4,h*0.9]);
        text3d(
            content,
            h=h,
            size=text_size,
            font=font,
            spacing=spacing,
            center=true,
            spin=spin,
            orient=orient
        );
        }
    } else{
    
    text3d(
        content,
        h=h,
        size=text_size,
        font=font,
        spacing=spacing,
        center=center,
        spin=spin,
        orient=orient
    );
    }
}

text_simple(
    content=str(content),
    text_size=text_size,
    h=h,
    font=font,
    spacing=spacing,
    center=center,
    spin=spin,
    orient=orient,
    hasBase=hasBase,
    baseHeight=baseHeight
);
